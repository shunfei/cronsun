# cronsun

`cronsun` 是一个分布式任务系统，单个结点和 `*nix` 机器上的 `crontab` 近似。支持界面管理机器上的任务，支持任务失败邮件提醒，安装简单，使用方便，是替换 `crontab` 一个不错的选择。

## 架构

```
                                                [web]
                                                  |
                                     --------------------------
           (add/del/update/exec jobs)|                        |(query job exec result)
                                   [etcd]                 [mongodb]
                                     |                        ^
                            --------------------              |
                            |        |         |              |
                         [node.1]  [node.2]  [node.n]         |
             (job exec fail)|        |         |              |
          [send mail]<-----------------------------------------(job exec result)

```

## Getting started

### Building the source

```
cd $GOPATH/src
git clone https://github.com/shunfei/cronsun.git
cd cronsun
sh ./build.sh
```

执行文件和配置文件在 `dist` 文件夹

### Run

1. 安装 [MongoDB](http://docs.mongodb.org/manual/installation/)
2. 安装 [etcd](https://github.com/coreos/etcd)
3. 修改 `conf` 相关的配置
4. 在任务结点启动 `./node -conf conf/base.json`，在管理结点启动 `./web -conf conf/base.json`
5. 访问管理界面 `http://127.0.0.1:7079/ui/`

## Screenshot

**Brief**:

![](doc/img/brief.png)

**Exec result**:

![](doc/img/log.png)

**Job**:

![](doc/img/job.png)

![](doc/img/new_job.png)

**Node**:

![](doc/img/node.png)
