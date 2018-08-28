# assert-dep

This command-line tool enables you to validate inter-packages dependencies within your golang project.

# Install

```shell
go get github.com/tomoemon/assert-dep
```

Ensure `$GOPATH` is set within your shell and your project is located in `$GOPATH`.

# Example usage

First, prepare a dependency rule file like below

https://github.com/tomoemon/assert-dep/blob/master/sampleProject/dep_rule_success.toml

Following command will succeed
```shell
assert-dep -config sampleProject/dep_rule_success.toml -root github.com/tomoemon/assert-dep/sampleProject
```
![sample_success.png](https://raw.githubusercontent.com/tomoemon/assert-dep/master/docs/sample_success.png)

Following command will fail (exit status: 1)
```shell
assert-dep -config sampleProject/dep_rule_error.toml -root github.com/tomoemon/assert-dep/sampleProject
```
![sample_error.png](https://raw.githubusercontent.com/tomoemon/assert-dep/master/docs/sample_error.png)
