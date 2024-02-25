package gee

import (
	"html/template"
	"net/http"
	"strings"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup     // store all groups
	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render
}

// New is the constructor of gee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup} // 初始化顶层 RG，没有进行实际分组
	return engine
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// ServeHTTP implements the http.Handler interface, to handle every HTTP request
// 通过实现 http.Handler 接口，Engine 就可以作为一个 HTTP 服务端来处理请求，代替原来的单独 Handler 处理
// 算是 `net/http` 给我们原生留下的起始扩展点
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups { // 直接对应的 Handler 在 `handle` 函数中取，这里只处理中间件
		if strings.HasPrefix(req.URL.Path, group.prefix) { // 这里说明一个实际 URL 可能匹配到多个 RG 多组中间件
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

// SetFuncMap 自定义模板渲染函数
// 用例比如说你需要使用一个有占位符的 html 模板，加上一个简单表示时间的 `gin.H` 来进行一个渲染
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
