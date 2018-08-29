# impas

impas is an IMPort ASsertion tool.
This command-line tool enables you to validate inter-packages dependencies within your golang project.

Most of practical projects consist of some kinds of layers, like UI, SERVICE, DAO, INFRA, ... etc. There are no problems if you develop the project by yourself because you should understand the whole project. On the other hand, team development may have some problems about understanding package dependency rules of the project. Especially, a new developer who doesn't know the whole project tends to write a code which breach the rule, because the rules are often implicit.

impas will make inter-packages dependency rules explicit.

# Install

```shell
go get github.com/tomoemon/impas
```

Ensure `$GOPATH` is set within your shell and your project is located in `$GOPATH`.

# Usage examples

First, prepare a dependency rule file like below

https://github.com/tomoemon/impas/blob/master/sampleProject/dep_rule_success.toml

Following command will succeed
```shell
impas -config sampleProject/dep_rule_success.toml -root github.com/tomoemon/impas/sampleProject
```
![sample_success.png](https://raw.githubusercontent.com/tomoemon/impas/master/docs/sample_success.png)

Following command will fail (exit status: 1)
```shell
impas -config sampleProject/dep_rule_error.toml -root github.com/tomoemon/impas/sampleProject
```
![sample_error.png](https://raw.githubusercontent.com/tomoemon/impas/master/docs/sample_error.png)
