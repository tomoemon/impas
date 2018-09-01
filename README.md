# impas

impas is an IMPort ASsertion tool.
This command-line tool enables you to validate inter-packages dependencies within your golang project.

Most of practical projects consist of some kinds of layers, like UI, SERVICE, DAO, INFRA, ... etc. There are no problems if you develop the project by yourself because you should understand the whole project. On the other hand, team development may have some problems about understanding package dependency rules of the project. Especially, a new developer who doesn't know the whole project tends to write a code which breach the rule, because such rules are often implicit.

Now, impas makes it explict!

<img src="https://raw.githubusercontent.com/tomoemon/impas/master/docs/images/impas.png" width=448>

# Prerequisites

- Using [dep](https://github.com/golang/dep) vendoring
- Your project is located in `$GOPATH`

# Install

```shell
go get -u github.com/tomoemon/impas
```


# Usage examples

First, prepare a dependency rule file like below

https://github.com/tomoemon/impas/blob/master/docs/sample_config_success.toml

Run following command
```shell
impas -config docs/sample_config_success.toml
```
<img src="https://raw.githubusercontent.com/tomoemon/impas/master/docs/images/sample_succeeded_result.png" width=448>

If command fail, it returns `exit status 1`
```shell
impas -config docs/sample_config_error.toml
```
<img src="https://raw.githubusercontent.com/tomoemon/impas/master/docs/images/sample_failed_result.png" width=600>
