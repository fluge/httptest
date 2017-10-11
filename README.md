# httptest

用于模拟HTTP请求返回对应的response

## Usage
配置host.test中的host，把请求转发到本地，如果不修改smart-dispatch中configs下的test.conf
中的文件，可以在本地起一个`80`端口的httptest服务。  
启动服务，修改test.conf中的端口号，然后go run httptestServer.go就行了

就下来就需要在json文件夹下编写对应的fixtures文件，
