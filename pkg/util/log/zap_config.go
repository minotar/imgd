package log

import "go.uber.org/zap"

func NewZapProd() (*zap.Logger, error) {
	zapConf := zap.NewProductionConfig()

	zapConf.Encoding = "console"
	zapConf.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	return zapConf.Build()
}
