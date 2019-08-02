# Simple image server 为简单而生  

sis更新到2.0，代码结构做了较大改动，主要实现以下两个功能：  
1. 实现一个简单的cache功能，cache空间填满后以先进先出原则管理cache 
2. 实现负载均衡功能，主程序可以以server和agent两种模式启动，可在agent端实现cache和文件md5运算  

#### 简易使用指南：  
1. 准备两台拥有独立IP的物理机或虚拟机，假定server IP为192.168.78.128,agent IP为192.168.78.130  
2. 在两台机器上安装好golang并编译sis  
3. 在192.168.78.128上的sis程序以server模式启动：./sis -port 4444 -cache 0
4. 在192.168.78.130上的sis程序以agent模式启动：./sis -localStore=false -image="http://192.168.78.128:4444"
5. 在agent端进入test/client目录，执行测试：go test -v
6. 测试通过后程序启动完成，此时通过agent上传的图片将会cache在agent的内存中，存储到server的硬盘里。

#### 关于docker
官方仓库已上传一份打包好的sis镜像，可以直接默认参数启动：  
>docker run -p 3333:3333 -d dhax/sis:v2.0

当然，也可以传递自定义参数，以server模式启动：  
>docker run -p 4444:4444 -d dhax/sis:v2.0 -port 4444 -cache 0

此时打开浏览器访问docker宿主机的相应端口，即可看到上传测试页面了

#### 关于k8s和sis3.0

sis3.0的目标是实现分布式存储，最初的想法是使用文件md5的第一个或前两个字符做服务器路由，agent自动找到后端相应的存储服务，但是要真正实现一个稳定可用的分布式存储工作量可能会非常巨大，不寒而栗。现在有了k8s，似乎什么都不用干了，真是懒人的福音啊，程序员可以下班了，一切交给运维吧:)。思路是这样的：k8s中sis以agent方式运行多副本，用来做负载均衡；以server方式运行单副本作为存储服务，然后给server的pod配置一个提供分布式存储的PVC，比如ceph什么的，这样就是一个3.0版的sis了，一切都很easy是不是，只是可能运维配这套东西要哭了吧:(。  
k8s运行 agent：
>kubectl run sis --image=index.docker.io/dhax/sis:v2.0 --port=3333 -- -localStore=false -image="http://*serverip:serverport*"  

k8s运行 server:  
>kubectl run sis --image=index.docker.io/dhax/sis:v2.0 --port=4444 -- -port 4444 -cache 0

当然这只是跑了一个deployment，一整套要跑起来都是k8s的内容了，完整的包含ceph的配置过程可能会非常长，有需要的去研究研究[Kubernetes](https://kubernetes.io/)吧。  

2019/6/3的华丽分割线
***

1.0版已诞生，除去graphics包，主体程序只有469行，版本特性：        
1. 实现文件上传接口，可支持一次提交多个文件。上传完成后服务器将返回json格式的数组，包含原始文件名和对应的MD5码   
2. 实现根据MD5码下载接口。如果发生MD5碰撞，此接口将返回找到的第一张图          
3. 实现根据MD5码和原始文件名下载接口    
4. 实现各下载接口对应的缩放接口           

#### 关于原始文件名：    
sis上传下载都可带上原始文件名，如此实现的目的是防止MD5碰撞的发生，即使两副图像的MD5码相同，只要原始文件名不同，也不会发生冲突   

#### 关于graphics包：    
由于官方发行包中暂无实现图像缩放功能的包，找了个半官方版本阉割后直接内置到sis中（https://code.google.com/archive/p/graphics-go/)

#### 简易使用指南：    
1. 下载安装golang(https://golang.google.cn/)     
1. [下载sis1.0源码](https://github.com/DDHax/sis/archive/1.0.1.tar.gz)
1. 将源码解压到目录： $HOME/go/src/github.com/DDHax/sis
1. *cd $HOME/go/src/github.com/DDHax/sis*
4. *go build sis.go*  
5. *nohup ./sis &*

此时服务已启动，可以使用sis test模块测试每个接口：   
*cd test/client/*  
*go test -v*   
全部PASS则说明sis已经在正常工作啦  
你也可以用浏览器访问主机的3333端口，默认主页有个简单的上传测试页面 

2018/8/7的华丽分割线
***         

sis背景：         
如今的互联网时代图片存储服务随处可见，实现方案也是五花八门，那么有没有一个开袋即食的方案呢？粗略找了一圈，[zimg](https://github.com/buaazp/zimg) 似乎是我最想要的，但一看长长的依赖安装列表顿时望而却步，虽然开袋即可吃了，但这袋子也太难开了点，手撕牙咬都不行，感觉要上剪刀。
于是sis诞生了，如果你也有这需求，赶紧拿走，别无他求，给加个星吧。

sis宪法：           
1. 程序安装不需前置依赖     
2. 程序编译不需前置依赖        
3. 程序启动不需配置文件          

sis实现：      
为了遵守宪法，似乎用GO实现是最好的选择。预计实现这么一个简单功能不会需要多少代码，那么开始吧。。。。。。      
上传接口：使用HTTP post       
下载接口：使用HTTP get       
文件存储：使用文件的MD5码拆解后作为目录名，文件原始文件存储在src目录，缩放后的文件根据尺寸单独建目录         
sis 1.0 将要实现的功能：       
- 图片上传     
- 图片下载     
- 图片缩放        

sis未来：     
cache有木有？可以有，但现在木有，2.0吧      
分布式有木有？可以有，但现在木有，3.0吧            
那4.0干啥呢？山还不够多么，不会有事了。      

