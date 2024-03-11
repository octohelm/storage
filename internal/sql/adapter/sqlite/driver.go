package sqlite

import (
	"context"
	"database/sql/driver"
	"io"
	"sync"
)

var _ driver.DriverContext = &driverContextWithMutex{}
var _ driver.Driver = &driverContextWithMutex{}

type driverContextWithMutex struct {
	driver.DriverContext
	*sync.Mutex
}

func (c *driverContextWithMutex) Driver() driver.Driver {
	return c
}

func (c *driverContextWithMutex) Open(name string) (driver.Conn, error) {
	cc, err := c.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return cc.Connect(context.Background())
}

func (c *driverContextWithMutex) OpenConnector(name string) (driver.Connector, error) {
	cc, err := c.DriverContext.OpenConnector(name)
	if err != nil {
		return nil, nil
	}
	return &connectorWithMutex{
		driverContextWithMutex: c,
		Connector:              cc,
	}, nil
}

var _ driver.Connector = &connectorWithMutex{}
var _ io.Closer = &connectorWithMutex{}

type connectorWithMutex struct {
	*driverContextWithMutex
	Connector driver.Connector
}

func (c *connectorWithMutex) Close() error {
	if cc, ok := c.Connector.(io.Closer); ok {
		return cc.Close()
	}
	return nil
}

func (c *connectorWithMutex) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &connWithMutex{Mutex: c.Mutex, Conn: conn}, nil
}

var _ driver.QueryerContext = &connWithMutex{}
var _ driver.ExecerContext = &connWithMutex{}
var _ driver.ConnBeginTx = &connWithMutex{}

type connWithMutex struct {
	*sync.Mutex
	driver.Conn
}

func (c *connWithMutex) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.Lock()
	defer c.Unlock()

	return c.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
}

func (c *connWithMutex) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.Lock()
	defer c.Unlock()

	return c.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
}

func (c *connWithMutex) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	c.Lock()
	defer c.Unlock()

	tx, err := c.Conn.(driver.ConnBeginTx).BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}

	return &txWithMutex{Mutex: c.Mutex, tx: tx}, nil
}

var _ driver.QueryerContext = &txWithMutex{}
var _ driver.ExecerContext = &txWithMutex{}

type txWithMutex struct {
	*sync.Mutex

	tx driver.Tx
}

func (c *txWithMutex) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.tx.(driver.ExecerContext).ExecContext(ctx, query, args)
}

func (c *txWithMutex) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.tx.(driver.QueryerContext).QueryContext(ctx, query, args)
}

func (c *txWithMutex) Commit() error {
	c.Lock()
	defer c.Unlock()

	return c.tx.Commit()
}

func (c *txWithMutex) Rollback() error {
	c.Lock()
	defer c.Unlock()

	return c.tx.Rollback()
}
