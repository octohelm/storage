package directive

type Directive struct {
	Args []any
	Name string
}

func (r *Directive) UnmarshalDirective(dec *Decoder) error {
	dd := Directive{}

	name, err := dec.DirectiveName()
	if err != nil {
		return err
	}
	dd.Name = name

	for {
		k, text := dec.Next()
		if k == EOF || k == KindFuncEnd {
			break
		}

		switch k {
		case KindValue:
			dd.Args = append(dd.Args, RawValue(text))
		case KindFuncStart:
			sub, err := dec.Unmarshaler(string(text))
			if err != nil {
				return err
			}
			if err := sub.UnmarshalDirective(dec); err != nil {
				return err
			}
			dd.Args = append(dd.Args, sub)
		default:

		}
	}

	*r = dd

	return nil
}

func (f Directive) MarshalDirective() ([]byte, error) {
	return MarshalDirective(f.Name, f.Args...)
}
