# 列出仓库级可用入口。
[group('meta')]
default:
    @just --list --unsorted

mod go 'tool/go'

# 启动测试依赖数据库。
[group('runtime')]
serve-dbs:
    cd hack && docker compose up -d
