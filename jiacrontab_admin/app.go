package admin

import (
	"errors"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"net/url"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/iris-contrib/middleware/cors"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
)

func newApp() *iris.Application {

	app := iris.New()
	app.UseGlobal(newRecover())
	app.Logger().SetLevel("debug")
	app.Use(logger.New())

	app.StaticEmbeddedGzip("/", "./assets/", GzipAsset, GzipAssetNames)
	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Jwt.SigningKey), nil
		},

		Extractor: func(ctx iris.Context) (string, error) {
			token, err := url.QueryUnescape(ctx.GetHeader(cfg.Jwt.Name))
			return token, err
		},
		Expiration: true,

		ErrorHandler: func(c iris.Context, data string) {
			ctx := wrapCtx(c)
			if ctx.RequestPath(true) != "/user/login" && ctx.RequestPath(true) != "/user/signUp" {
				ctx.respAuthFailed(errors.New("认证失败"))
				return
			}
			ctx.Next()
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	crs := cors.New(cors.Options{
		Debug:            true,
		AllowedHeaders:   []string{"Content-Type", "Token"},
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
	})
	app.Use(crs)
	app.AllowMethods(iris.MethodOptions)
	app.Get("/", func(ctx iris.Context) {
		asset, _ := GzipAsset("index.html")
		ctx.HTML(string(asset))
	})
	adm := app.Party("/adm")
	{
		adm.Use(jwtHandler.Serve)
		adm.Post("/crontab/job/list", GetJobList)
		adm.Post("/crontab/job/get", GetJob)
		adm.Post("/crontab/job/log", GetRecentLog)
		adm.Post("/crontab/job/edit", EditJob)
		adm.Post("/crontab/job/action", ActionTask)
		adm.Post("/crontab/job/exec", ExecTask)

		adm.Post("/config/get", GetConfig)
		adm.Post("/config/mail/send", SendTestMail)
		adm.Post("/system/info", SystemInfo)

		adm.Post("/daemon/job/list", GetDaemonJobList)
		adm.Post("/daemon/job/action", ActionDaemonTask)
		adm.Post("/daemon/job/edit", EditDaemonJob)
		adm.Post("/daemon/job/get", GetDaemonJob)
		adm.Post("/daemon/job/log", GetRecentDaemonLog)

		adm.Post("/group/list", GetGroupList)
		adm.Post("/group/edit", EditGroup)

		adm.Post("/node/list", GetNodeList)
		adm.Post("/node/delete", DeleteNode)
		adm.Post("/node/group_node", GroupNode)

		adm.Post("/user/activity_list", GetActivityList)
		adm.Post("/user/job_history", GetJobHistory)
		adm.Post("/user/audit_job", AuditJob)
		adm.Post("/user/stat", UserStat)
		adm.Post("/user/signup", Signup)
		adm.Post("/user/edit", EditUser)
		adm.Post("/user/group_user", GroupUser)
		adm.Post("/user/list", GetUserList)
	}

	app.Post("/user/login", Login)
	app.Post("/app/init", InitApp)

	debug := app.Party("/debug")
	{
		debug.Get("/stat", stat)
		debug.Get("/pprof/", indexDebug)
		debug.Get("/pprof/{key:string}", pprofHandler)
	}

	return app
}
