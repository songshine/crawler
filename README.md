# 网页数据爬取工具

## 核心组件

- Pager：根据分页URL请求格式， 获取某一范围的所有页的response
- Ruler： 指定网页response 分析规则
- URL Collector：　依赖`Pager` 和`Ruler`　收集所有的需要最终爬取数据的页面的URL集合
- Data Collector: 从`URL Collector` 中读取URL， 并指定`Ruler` 集合， 让后爬取相关数据
- Data Storage: 从`Data Collector` 中读取数据存储到指定位置， 现在只支持到CSV

