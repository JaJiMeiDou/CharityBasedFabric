# CharityBasedFabric
fabric, blockchain ,charity platform

1 概述

1.1项目背景

近期由于新冠病毒疫情爆发，在假期返乡潮期间，部分地区防疫物资出现临时短缺的情况。信息一经披露，各地纷纷献出援手，支援各地医院的抗击疫情活动。围绕抗“疫”物资援助和分配的问题，某红十字会陷入了舆论的旋涡，其拨付缓慢、物资分配不公、手续复杂、不公开透明的问题引起了大众的关心与重视。现今社会的公益体系下的信任危机，正在吞噬人们爱心慈善的热情。
依赖区块链技术，可以使需要帮助的人能及时发布求助信息，使善良的人能有的放矢，及时奉献自己的爱心，使得每个善举都公开透明，有迹可循，不可篡改，激发社会正能量，助力社会公益，为祖国的公益事业添砖加瓦。

1.2项目简介

该项目旨在构建一个具有帮助用户发布和了解求助信息，记录爱心善举和捐赠物资流向等功能的公益平台。该平台采用BS架构，通过利用fabric底层区块链相关技术和接口，将用户发布的求助和援助信息，以及合作物流企业方面提供的物资流动信息上链，使信息公开透明，不可篡改；利用二维码等技术确保物资的流向真实有效，可追溯；利用IPFS相关技术便于图片上链；利用bootstrap等前端技术与框架，方便用户的操作使用。


2 运行环境

2.1硬件要求

CPU：5代以上intel处理器，频率高于1.5GHz。

内存：4G及以上。

硬盘空间：50G以上。

网络：50M带宽。

2.2软件环境要求

操作系统：推荐使用unix及其衍生的操作系统，如MacOS、Linux。 

编译环境：Golang 1.12及以上。

其他环境：Docker 19.01.0及以上、Docker-Compose 1.25.0及以上IPFS 0.5及以上。

开发环境：visual studio code。


3 后端初始化

3.1fabric网络配置

在正确设置好上述运行环境后，后从github上获取fabric的源码，存放在本地，如存放路径为“~/go/src/github.com/hyperledger/fabric”，然后使用git checkout命令将fabric的版本切换到1.0.0，然后下载运行Fabric所必需的Docker镜像。

在fabric文件夹下，新建文件夹CharitySystem，这便是整个物资援助溯源平台的目录。进入此目录中，再新建一个deploy目录，作为有关系统网络配置信息的目录，再进入这个目录中。

首先根据自己的需求修改配置文件模板crypto-config.yaml，这个模板的作用是配置peer节点组织和orderer节点组织的相关信息，如组织名称和域名。然后使用命令cryptogen generate --config=crypto-config.yaml，生成对应的组织的节点的用户证书，证书会存放在cryto-config文件夹中。

然后编写配置文件configtx.yaml，这个配置文件指定了排序节点的共识策略，共识策略可以为Kafka；此外还可以配置与出块有关的信息，如出块时间BatchTimeout，一个区块中允许的最大交易数MaxMessageCount等。

然后使用命令configtxgen -profile生成创世区块和通道的创世交易，然后生成组织关于通道的主节点交易。这些生成的文件都存放在config目录下。

之后需要编写docker-compose配置文件，目的是生成客户端运行容器，节点运行容器，指定这些容器暴露的端口等。与docker有关的配置文件编写完毕后，在deploy目录下使用命令docker-compose up -d启动容器。
然后围绕已经创建的peer节点，需要使用命令peer channel create …和命令peer channel join …创建并加入通道，然后通过命令peer channel update…设置主节点。

之后，将平台代码分别存放在正确的目录，将链码安装在peer节点上并实例化，即可运行链码。系统目录中，与前端相关的代码存放到了目录application中，其中view目录存放各个html文件；reso目录存放它们用到的一些资源，包括js代码，字体，图片等；main.go文件是基于gin框架的web服务代码，已在平台代码文件中给出，编译出的二进制文件是application，config.yaml是它的一些配置信息。

3.2启动后端服务

	使用docker-compose up -d启动fabric网络容器，使用命令ipfs daemon启动IPFS节点服务器，进入application目录，使用命令go build编译，使用./application 命令运行即可启动所有后端服务。
