// 链路追踪

package tracex

import (
	"github.com/climber-dong/global/genv"
	"github.com/opentracing/opentracing-go"
	"log"
)

const (
	TraceTypeJaeger = "jaeger"
	TraceTypeZipkin = "zipkin"
)

// tracer 全局单例变量
var tracer opentracing.Tracer

// InitTrace 初始化trace
func getTracer() (opentracing.Tracer, error) {
	if tracer == nil {
		if genv.TraceMod() != 0 {
			var err error
			switch genv.TraceType() {
			case TraceTypeZipkin:
				tracer = newZkTracer(genv.AppName(), genv.LocalIP(), genv.TraceAddr(), genv.TraceMod())
				if err != nil {
					return nil, err
				}
				log.Printf("[app] tracer success")
			case TraceTypeJaeger:
				tracer = newJTracer(genv.AppName(), genv.TraceAddr(), genv.TraceMod())
				if err != nil {
					return nil, err
				}
				log.Printf("[app] tracer success")
			}
		}
	}

	return tracer, nil
}
