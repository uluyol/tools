package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
)

func defaultConfigDir() string {
	return filepath.Join(xdgConfigDir(), "remote-do")
}

func xdgConfigDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return d
	}
	var h string
	if u, err := user.Current(); err == nil {
		h = u.HomeDir
	} else {
		h = os.Getenv("HOME")
	}
	return filepath.Join(h, ".config")
}

var (
	configDir      = flag.String("configdir", defaultConfigDir(), "path to configuration directory")
	selectedRemote = flag.String("remote", "", "remote server to use (defaults to default in config)")
	namePre        = flag.String("name", "", "prefix to give session file")
)

func usage() {
	fmt.Fprintln(os.Stderr, "remote-do [args] cmd args...")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetPrefix("remote-do: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
	}
	config, err := loadConfig(*configDir)
	if err != nil {
		log.Fatal(err)
	}
	remoteName := *selectedRemote
	if remoteName == "" {
		remoteName = config.DefaultRemote
	}
	if remoteName == "" {
		log.Fatal("no selected or default remote")
	}
	remote, ok := config.Remotes[remoteName]
	if !ok {
		log.Fatalf("selected remote %q has no config", remoteName)
	}
	if hasRelativeRemotePaths(remote) {
		remHome, err := getRemoteHome(remote)
		if err != nil {
			log.Fatalf("unable to get remote home directory: %v", err)
		}
		remote = resolveRelative(remote, remHome)
	}
	pre := ""
	if *namePre != "" {
		pre = *namePre + "-"
	}
	session, f, err := mkSessionFile(pre)
	if err != nil {
		log.Fatal(err)
	}

	remCacheDir := path.Join(remote.TopDir, "cache")
	remSessionDir := path.Join(remote.TopDir, session)
	remTempDir := remSessionDir + "-tmp"

	fmt.Println("clearing data and copying files...")
	_ = cexec("ssh", remote.Host, fmt.Sprintf(
		"rm -rf '%s'; [[ -d '%s' ]] && cp -R '%s' '%s'",
		remSessionDir, remCacheDir, remCacheDir, remSessionDir))
	args := []string{"-az", "--delete", "."}
	if _, err := os.Stat(".remote-do-ignore"); err == nil {
		args = []string{"-az", "--delete", "--exclude-from=.remote-do-ignore", "."}
	}
	err = cexec("rsync", append(args, remote.Host+":"+remSessionDir+"/")...)
	if err != nil {
		log.Fatalf("failed to copy files: %v", err)
	}
	err = cexec("ssh", remote.Host,
		fmt.Sprintf("rsync -a --delete '%s/' '%s'", remSessionDir, remCacheDir))
	if err != nil {
		log.Fatalf("failed to update remote cache: %v", err)
	}

	fmt.Println("running commands...")
	doRun, err := execTmpl(doRunTmpl, struct {
		Cmd []string
	}{
		Cmd: flag.Args(),
	})
	if err != nil {
		log.Fatalf("unable to execute doRun template: %v", err)
	}
	if !strings.HasPrefix(remote.JunestHome, "/") {
		remote.JunestHome = "$HOME/" + remote.JunestHome
	}
	if !strings.HasPrefix(remote.JunestRepo, "/") {
		remote.JunestRepo = "$HOME/" + remote.JunestRepo
	}
	scriptBody, err := execTmpl(scriptBodyTmpl, struct {
		TempDir    string
		JunestCmd  []string
		SessionDir string
	}{
		TempDir:    remTempDir,
		JunestCmd:  []string{"JUNEST_HOME=" + remote.JunestHome, path.Join(remote.JunestRepo, "bin", "junest")},
		SessionDir: remSessionDir,
	})
	if err != nil {
		log.Fatalf("unable to execute scriptBody template: %v", err)
	}
	cmdRun, err := execTmpl(cmdRunTmpl, struct {
		DoRun      string
		ScriptBody string
		SessionDir string
	}{
		DoRun:      doRun,
		ScriptBody: scriptBody,
		SessionDir: remSessionDir,
	})
	if err != nil {
		log.Fatalf("unable to execute cmdRun template: %v", err)
	}

	if err := cexec("ssh", remote.Host, cmdRun); err != nil {
		log.Fatalf("error running command: %v", err)
	}

	err = sessionBodyTmpl.Execute(f, struct {
		Host       string
		SessionDir string
		TempDir    string
		Cmd        []string
	}{
		Host:       remote.Host,
		SessionDir: remSessionDir,
		TempDir:    remTempDir,
		Cmd:        flag.Args(),
	})
	if err != nil {
		log.Fatalf("unable to execute sessionBody template: %v", err)
	}

	f.Close()
	os.Chmod(session, 0755)
	fmt.Println(session)
}

func getRemoteHome(r Remote) (string, error) {
	out, err := exec.Command("ssh", r.Host, "echo", "$HOME").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func cexec(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func execTmpl(t *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	return buf.String(), err
}

const doRun = `#!/usr/bin/env bash
export TERM=xterm
export TMPDIR=/runtmp
cd "${0%/*}"
{{range $i, $e := .Cmd}}"{{$e}}" {{end}};
`

const scriptBody = `#!/usr/bin/env bash
mkdir -p "{{.TempDir}}"
env {{range $i, $e := .JunestCmd}}"{{$e}}" {{end}} -p "-b {{.TempDir}}:/runtmp" {{.SessionDir}}/remote-inner.bash
`

const cmdRun = `\
echo '{{.DoRun}}'>'{{.SessionDir}}/remote-inner.bash' && \
echo '{{.ScriptBody}}'>'{{.SessionDir}}/remote-outer.bash' && \
chmod a+x '{{.SessionDir}}/remote-inner.bash' && \
chmod a+x '{{.SessionDir}}/remote-outer.bash' && \
'{{.SessionDir}}/remote-outer.bash'
`

const sessionBody = `#!/usr/bin/env bash
if [[ $1 == clean ]]; then
	# clean will delete files for us that are in the home directory.
	# This lets us safely use root to do so without worrying about
	# deleting unwanted things.
	# This is required because of the permissions docker files are
	# written with.
	ssh "{{.Host}}" "rm -rf '{{.SessionDir}}' '{{.TempDir}}'"
	rm -f "$0"
elif [[ $1 == cmd ]]; then
	echo {{range $i, $e := .Cmd}}{{$e}} {{end}}
else
	echo "usage: $0 cmd|clean" >&2
	exit 1
fi
`

var (
	doRunTmpl       = template.Must(template.New("doRun").Parse(doRun))
	scriptBodyTmpl  = template.Must(template.New("scriptBody").Parse(scriptBody))
	cmdRunTmpl      = template.Must(template.New("cmdRun").Parse(cmdRun))
	sessionBodyTmpl = template.Must(template.New("sessionBody").Parse(sessionBody))
)

func mkSessionFile(namePre string) (string, *os.File, error) {
Start:
	session := namePre + "session." + strconv.FormatInt(rand.Int63(), 36)[:5]
	f, err := os.OpenFile(session, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		if os.IsExist(err) {
			goto Start
		}
	}
	return session, f, err
}

const configTmpl = `
# Example config file (formatted using TOML).
# You should set default-remote and it's config.
#
#default-remote = default
#
#[remotes.default]
#host = www.example.com
#top-dir = /path/to/dir # or path/to/dir to be relative to your home dir
#junest-repo = /path/to/dir
#junest-home = /path/to/dir
`

func resolveRelative(r Remote, home string) Remote {
	for _, sp := range [...]*string{&r.TopDir, &r.JunestRepo, &r.JunestHome} {
		if !strings.HasPrefix(*sp, "/") {
			*sp = path.Join(home, *sp)
		}
	}
	return r
}

func hasRelativeRemotePaths(r Remote) bool {
	for _, s := range [...]string{r.TopDir, r.JunestRepo, r.JunestHome} {
		if !strings.HasPrefix(s, "/") {
			return true
		}
	}
	return false
}

type Remote struct {
	Host       string `toml:"host"`
	TopDir     string `toml:"top-dir"`
	JunestRepo string `toml:"junest-repo"`
	JunestHome string `toml:"junest-home"`
}

type Config struct {
	DefaultRemote string            `toml:"default-remote"`
	Remotes       map[string]Remote `toml:"remotes"`
}

func loadConfig(d string) (*Config, error) {
	if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
		os.MkdirAll(d, 0755)
	}
	cp := filepath.Join(d, "config.toml")
	if _, err := os.Stat(cp); err != nil {
		ioutil.WriteFile(cp, []byte(configTmpl), 0644)
		return nil, fmt.Errorf("missing config, created skeleton at %s", cp)
	}

	var config Config
	_, err := toml.DecodeFile(cp, &config)
	return &config, err
}
