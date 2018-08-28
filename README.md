# assert-dep

This command-line tool enables you to validate inter-packages dependencies within your golang project.

# Install

```shell
go get github.com/tomoemon/assert-dep
```

Ensure `$GOPATH` is set within your shell and your project is located in `$GOPATH`.

# Example usage

Following command is successs example
```shell
assert-dep -config sampleProject/dep_rule_success.toml -root github.com/tomoemon/assert-dep/sampleProject
```

Following command is failure example
```shell
assert-dep -config sampleProject/dep_rule_error.toml -root github.com/tomoemon/assert-dep/sampleProject
```
