api-server整体设计思路:
    职责：
    业务系统（golang）管理所有api（runner）的地址包括api的请求参数，请求方式，响应参数，接口版本等等
    是存储runner的元数据，还有和runcher建立nats连接，因为api-server是在公有云部署，所以他本身必须包含nats服务，然后在私有云的runcher需要带服务发现功能，自动注册连接公有云的api-server，

运行流程：
    前端在web界面发送请求到api-server，然后api-server通过查询这个接口获取到runner信息，然后我们可以根据runner信息给runcher发送请求，
    携带前端的请求信息给runcher，然后runcher处理后返回给我们，期间api-server还需要包含管理runner的升级，添加api，添加package，runner的创建，
    runner的删除，runner的回滚，等等，这些都是apiserver和runcher之间的交互，然后runcher再调度runner，然后把消息返回到api-server，
    api-server的职责不负责计算和执行程序逻辑，只负责请求的加工然后转发给runcher，然后api-server需要存储当前runner的版本，所属用户，api等等元数据相关信息

模块规划
    runner 表，runner在前端的表现是一个项目（一个项目下有服务目录，服务目录下有服务目录和运行函数）
        存储用户的runner的元数据
        存储runner的所属用户，名称，部署在哪个runcher机器上，是用什么sdk编译的，当前的版本是多少，这runner的标签，等等
        提供runner的管理接口，包括runner的创建，runner的删除，runner的修改，runner的回滚，等等
        提供runner的查询接口，包括runner的查询，runner的查询列表，等等
        提供runner的发布历史接口，每次发布都记录发布的版本，发布的时间，发布的用户，等等
    runner_func 表
        sdk-go编译后的runner都自带一个/_getApiInfos的路由，这个路由返回runner的所有接口信息，包括接口的名称，接口的标签，接口的请求参数，接口的响应参数，接口的回调列表，接口用到的tables，等等，每次发布可以调用这个接口把相关的api同步到runner_func表中，这样就可以实现runner的自动注册和管理，
    service_tree表，是相当于项目里的目录，相当于go的package，里面可以放函数或者继续嵌套package
        func是相当于项目里的函数，相当于go的函数
        runner_func表是用来存储runner的所有接口信息的，包括接口的名称，接口的标签，接口的请求参数，接口的响应参数，接口的回调列表，
        接口用到的tables，等等，每次发布可以调用这个接口把相关的api同步到runner_func表中，这样就可以实现runner的自动注册和管理，
        项目可以基于package进行fork，比如清华大学有个清华常用函数库，然后他们没有数学相关的函数库，
        去fork了北大的北大常用函数库下的数学库到清华常用函数库下面，这时候清华大学常用函数库下面就有了个数学相关的函数库，
        然后清华和北大他们就可以快快乐乐的互相用自己的数学库了，互不影响，各自之间的数据和结构都是独立的，
        清华在自己的数学库下面再加和删除func或者package都不会影响北大的库，后面北大的数学库下面多了很多个函数，
        清华想用其中一个求斐波那契数列的函数，然后清华可以单独fork这个函数到清华的数学库也是支持的，就像git的fork功能一样互相独立，
        又可以互相融合，这时候就需要一个runner的表来存储runner的信息，然后一个service_tree的表来存储服务树的信息，
        然后一个func的表来存储func的信息（包含请求参数和响应参数），这样就可以实现runner的自动注册和管理，说一下这个fork的原理，
        sdk生成的代码都是一个函数一个go文件，一个service_tree对应一个go的package，所以fork的原理非常简单，
        就是把runcher会把北大的函数库直接复制到清华的函数库的指定位置，然后重新编译清华的runner并升级版本即可