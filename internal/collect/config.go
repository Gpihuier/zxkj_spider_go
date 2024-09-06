package collect

import (
	"fmt"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/gookit/slog"
)

type Config struct {
	Name      string   `mapstructure:"name"`
	DomainUrl string   `mapstructure:"domain_url"`
	Soft      Template `mapstructure:"soft"`
	News      Template `mapstructure:"news"`
}

type Template struct {
	Urls []string `mapstructure:"urls"`
	List struct {
		Node string `mapstructure:"node"`
		Href string `mapstructure:"href"`
	} `mapstructure:"list"`
	Content map[string]string  `mapstructure:"content"`
	Process map[string]Process `mapstructure:"process,omitempty"`
}

type Process struct {
	Trim    []string          `mapstructure:"trim,omitempty"`
	Replace map[string]string `mapstructure:"replace,omitempty"`
}

func LoadTasks() <-chan *Config {
	ch := make(chan *Config)

	go func() {
		defer func() {
			close(ch)
		}()

		config.WithDriver(yaml.Driver)

		// 获取任务列表
		err := config.LoadFiles("../../config/collect/tasks.yaml")
		if err != nil {
			slog.Error(err)
			return
		}
		tasks := config.Strings("list")

		// 解析任务配置文件
		for _, task := range tasks {
			config.ClearAll()
			if err = config.LoadFiles(fmt.Sprintf("../../config/collect/%s.yaml", task)); err != nil {
				slog.Error(err)
				return
			}
			var template Config
			if err = config.Decode(&template); err != nil {
				slog.Error(err)
				return
			}
			if template.Name == "" {
				template.Name = task
			}

			ch <- &template
		}

	}()

	return ch
}
