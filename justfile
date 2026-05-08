# 列出所有可用命令（含子模块，无输入）
[group('meta')]
default:
    @just --list --list-submodules

[group('env')]
serve-dbs:
    cd hack && docker compose up -d

# Go 工具链入口
[group: 'toolchain']
mod go 'tool/go'
