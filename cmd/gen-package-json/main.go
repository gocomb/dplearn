package main

import (
	"bytes"
	"flag"
	"os"
	"text/template"
	"time"

	"strings"

	"github.com/golang/glog"
	"github.com/gyuho/deephardway/pkg/gcp"
)

func main() {
	configPath := flag.String("config", "package.json", "Specify config file path.")
	flag.Parse()

	// TODO: angular-cli does not work with public IP, so need to use 0.0.0.0
	// https://github.com/angular/angular-cli/issues/2587#issuecomment-252586913
	// but webpack prod mode does not work with 0.0.0.0
	// so just use without prod mode
	// https://github.com/webpack/webpack-dev-server/issues/882
	cfg := configuration{
		NgCommandServeStart:     "ng serve",
		NgCommandServeStartProd: "ng serve --prod",
		Host:         "0.0.0.0",
		HostPort:     4200,
		HostProd:     "0.0.0.0",
		HostProdPort: 4200,
	}

	for i := 0; i < 3; i++ {
		// inspect metadata to get public IP
		bts, err := gcp.GetComputeMetadata("instance/network-interfaces/0/access-configs/0/external-ip")
		if err != nil {
			glog.Warning(err)
			time.Sleep(300 * time.Millisecond)
			continue
		}
		// TODO
		// cfg.HostProd = host
		glog.Infof("found public host IP %q", strings.TrimSpace(string(bts)))
		break
	}

	buf := new(bytes.Buffer)
	tp := template.Must(template.New("tmplPackageJSON").Parse(tmplPackageJSON))
	if err := tp.Execute(buf, &cfg); err != nil {
		glog.Fatal(err)
	}
	txt := buf.String()

	if err := toFile(txt, *configPath); err != nil {
		glog.Fatal(err)
	}
	glog.Infof("wrote %q", *configPath)
}

type configuration struct {
	NgCommandServeStart     string
	NgCommandServeStartProd string
	Host                    string
	HostPort                int
	HostProd                string
	HostProdPort            int
}

const tmplPackageJSON = `{
    "name": "app-deephardway",
    "version": "0.9.0",
    "license": "Apache-2.0",
    "angular-cli": {},
    "scripts": {
        "start": "{{.NgCommandServeStart}} --port {{.HostPort}} --host {{.Host}} --proxy-config proxy.config.json",
        "start-prod": "{{.NgCommandServeStartProd}} --port {{.HostProdPort}} --host {{.HostProd}} --disable-host-check --proxy-config proxy.config.json",
        "lint": "tslint \"frontend/**/*.ts\"",
        "test": "ng test",
        "pree2e": "webdriver-manager update",
        "e2e": "protractor"
    },
    "private": true,
    "dependencies": {
        "@angular/common": "4.1.3",
        "@angular/compiler": "4.1.3",
        "@angular/compiler-cli": "4.1.3",
        "@angular/core": "4.1.3",
        "@angular/forms": "4.1.3",
        "@angular/http": "4.1.3",
        "@angular/platform-browser": "4.1.3",
        "@angular/platform-browser-dynamic": "4.1.3",
        "@angular/animations": "4.1.3",
        "@angular/router": "4.1.3",
        "@angular/tsc-wrapped": "4.1.3",
        "@angular/upgrade": "4.1.3",
        "core-js": "2.4.1",
        "rxjs": "5.4.0",
        "ts-helpers": "1.1.2",
        "zone.js": "0.8.11"
    },
    "devDependencies": {
        "@angular/cli": "1.2.0-beta.0",
        "@types/angular": "1.6.18",
        "@types/angular-animate": "1.5.6",
        "@types/angular-cookies": "^1.4.2",
        "@types/angular-mocks": "1.5.9",
        "@types/angular-resource": "1.5.8",
        "@types/angular-route": "1.3.3",
        "@types/angular-sanitize": "1.3.4",
        "@angular/material": "2.0.0-beta.6",
        "@types/hammerjs": "2.0.34",
        "@types/jasmine": "2.5.51",
        "@types/node": "7.0.27",
        "codelyzer": "3.0.1",
        "jasmine-core": "2.6.2",
        "jasmine-spec-reporter": "4.1.0",
        "karma": "1.7.0",
        "karma-chrome-launcher": "2.1.1",
        "karma-cli": "1.0.1",
        "karma-jasmine": "1.1.0",
        "karma-remap-istanbul": "0.6.0",
        "protractor": "5.1.2",
        "typescript": "2.3.4",
        "ts-node": "3.0.4",
        "tslint": "5.4.2"
    },
    "description": "website",
    "main": "index.js",
    "repository": {
        "url": "git@github.com:gyuho/deephardway.git",
        "type": "git"
    },
    "author": "Gyu-Ho Lee <gyuhox@gmail.com>"
}
`

func toFile(txt, fpath string) error {
	f, err := os.OpenFile(fpath, os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		f, err = os.Create(fpath)
		if err != nil {
			glog.Fatal(err)
		}
	}
	defer f.Close()
	if _, err := f.WriteString(txt); err != nil {
		glog.Fatal(err)
	}
	return nil
}

// exist returns true if the file or directory exists.
func exist(fpath string) bool {
	st, err := os.Stat(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	if st.IsDir() {
		return true
	}
	if _, err := os.Stat(fpath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func nowPST() time.Time {
	tzone, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.Now()
	}
	return time.Now().In(tzone)
}
