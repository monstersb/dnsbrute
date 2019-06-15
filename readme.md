# DNS Brute

### 使用前先 build 一下

```
go build brute.go
```

### 使用方式

```
$ ./brute -t chaitin.cn
```

### 关于这个东西

和同事讨论到用 python 写 dns 爆破太慢，提了几个思路:

1. 用 go 优化并发
2. 实现简单的 dns 协议，替代系统提供的 dns 解析
3. 在同一个 dns 查询请求里尝试放入多条记录（but 这个由于 dns 放大攻击导致大部分 dns 服务器都禁用了，保留了代码，但是默认没启用）

测试结果:

在并发和带宽合理的情况下，每秒的查询数量在 1500 左右。   

```
$ ./brute -t chaitin.cn
Total:   17576
Result:  2
Error:   2 timeout
cost:    13 seconds
```

问题:

由于用 udp 直接实现了 dns 查询，因此丢包的现象比较严重，如果在意的话可以加上 retry 机制

### 参数说明

```
$ ./brute -h
Usage of ./brute:
  -a string
    	Brute Alphabet (default "abcdefghijklmnopqrstuvwxyz")
  -l int
    	Sub Domain Name Length (default 3)
  -o string
    	Output File (default "output.txt")
  -t string
    	Target You Want To Bruteforce
```
