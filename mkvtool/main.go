package main

import (
	"fmt"
	"github.com/MkvAutoSubset/MkvAutoSubset/mkvlib"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

const appName = "MKV Tool"
const appVer = "v5.0.6"
const tTitle = appName + " " + appVer

var appFN = fmt.Sprintf("%s %s %s/%s", appName, appVer, runtime.GOOS, runtime.GOARCH)

var latestTag = ""

var (
	ec    int
	nck   bool
	cache string
)

var processer = mkvlib.GetProcessorGetterInstance().GetProcessorInstance()

func main() {
	setWindowTitle(tTitle)
	go getLatestTag()

	defer func() {
		if latestTag != "" && latestTag != appVer {
			_os := strings.ToUpper(string(runtime.GOOS[0])) + runtime.GOOS[1:]
			arch := runtime.GOARCH
			if arch == "amd64" {
				arch = "x86_64"
			}
			ext := "tar.gz"
			if _os == "Windows" || _os == "Darwin" {
				ext = "zip"
			}
			color.Green("New version available: %s\nDownload link: https://github.com/MkvAutoSubset/MkvAutoSubset/releases/download/%s/mkvtool_%s_%s_%s.%s", latestTag, latestTag, latestTag[1:], _os, arch, ext)
		}
		os.Exit(ec)
	}()

	cmd := &cobra.Command{
		Use:   "mkvtool <subcommand>",
		Short: appName + " - A versatile tool for managing mkv files",
		Long:  appName + ` is a comprehensive utility for creating, dumping, making, querying, and managing mkv files and their associated subtitles and fonts.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cache = os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		cache = os.Getenv("USERPROFILE")
	}
	cache = path.Join(cache, ".mkvtool", "caches")

	flog := ""
	ncks := false
	n := false
	a2p := false
	apc := false
	pf := ""
	pr := ""

	cmd.PersistentFlags().StringVarP(&flog, "log", "", "", "Specify the log file path.")
	cmd.PersistentFlags().StringVarP(&cache, "font-cache-dir", "", cache, "Specify the font cache directory path.")
	cmd.PersistentFlags().BoolVarP(&nck, "no-font-check", "", false, "Disable font check mode.")
	cmd.PersistentFlags().BoolVarP(&ncks, "no-font-check-strict", "", false, "Disable strict font check mode.")
	cmd.PersistentFlags().BoolVarP(&n, "no-font-rename", "", false, "Disable font rename mode.")
	cmd.PersistentFlags().BoolVarP(&a2p, "enable-pgs-output", "", false, "Enable pgs output mode.")
	cmd.PersistentFlags().BoolVarP(&apc, "enable-ass-pgs-coexist", "", false, "Enable pgs and ass coexistence mode.")
	cmd.PersistentFlags().StringVarP(&pf, "framerate", "", "23.976", "Set pgs or blank video frame rate (e.g., 23.976, 24, 25, 30, 29.97, 50, 59.94, 60, or custom fps like 15/1).")
	cmd.PersistentFlags().StringVarP(&pr, "resolution", "", "1920*1080", "Set pgs or blank video resolution (e.g., 720p, 1080p, 2k, or custom resolution like 720*480).")

	cmd.AddCommand(versionCmd())
	cmd.AddCommand(infoCmd())
	cmd.AddCommand(listCmd())
	cmd.AddCommand(queryCmd())
	cmd.AddCommand(dumpCmd())
	cmd.AddCommand(makeCmd())
	cmd.AddCommand(createCmd())
	cmd.AddCommand(subsetCmd())
	cmd.AddCommand(cacheCmd())

	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if flog != "" {
			l, err := os.Create(flog)
			if err != nil {
				color.Red(`Failed to create log file: "%s".`, flog)
			}
			mw := io.MultiWriter(colorable.NewColorableStdout(), l)
			color.Output = mw
			color.NoColor = true
		}

		if !mkvlib.GetProcessorGetterInstance().InitProcessorInstance(nil) {
			ec++
			color.Red("Failed to init processor.")
			return
		} else {
			processer = mkvlib.GetProcessorGetterInstance().GetProcessorInstance()
			ccs, _ := findPath(cache, `\.cache$`)
			processer.Cache(ccs)
			processer.Check(!nck, !ncks)
			processer.A2P(a2p, apc, pr, pf)
			processer.NRename(n)
		}
	}

	if err := cmd.Execute(); err != nil {
		ec++
	}

	// _ = doc.GenMarkdownTree(cmd, "docs")
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Args:    cobra.NoArgs,
		Short:   "Display the application version",
		Long:    `Show the current version of mkvtool along with the platform information.`,
		Run: func(cmd *cobra.Command, args []string) {
			color.Green("%s (powered by %s)", appFN, mkvlib.LibFName)
		},
	}
}

func infoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info <font-file>",
		Aliases: []string{"i"},
		Short:   "Display font information",
		Long:    `Show detailed information about the fonts used in a specified file.`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			info := processer.GetFontInfo(args[0])
			if info != nil {
				color.Blue("File: \t%s\n", info.File)
				l := len(info.Fonts)
				for i := 0; i < l; i++ {
					color.Magenta("\nIndex:\t%d\n", i)
					color.Green("\tNames:\t%s\n", strings.Join(info.Fonts[i], "\n\t\t"))
					color.HiGreen("\tTypes:\t%s\n", strings.Join(info.Types[i], "\n\t\t"))
				}
			} else {
				color.Red("Failed to get font info: [%s].", args[0])
				ec++
			}
		},
	}

	return cmd
}

func listCmd() *cobra.Command {
	f := ""
	cmd := &cobra.Command{
		Use:     "list <ass-file|ass-dir> [output-dir]",
		Aliases: []string{"l"},
		Short:   "Show list of fonts",
		Args:    cobra.RangeArgs(1, 2),
		Long:    `Display a list of fonts used in the specified ass file or folder.`,
		Run: func(cmd *cobra.Command, args []string) {
			files := []string{args[0]}
			if i, _ := os.Stat(files[0]); i.IsDir() {
				files, _ = findPath(files[0], `\.ass`)
			}
			list := processer.GetFontsList(files, f, nil)
			if len(list[0]) > 0 {
				color.Yellow("Required fonts: \t%s\n", strings.Join(list[0], "\n\t\t"))
				if len(list[1]) > 0 {
					color.HiYellow("\nMissing fonts: \t%s\n", strings.Join(list[1], "\n\t\t"))
				} else if !nck {
					color.Green("\n*** All required fonts are present. ***")
				}
			} else {
				color.Yellow("!!! No fonts found. !!!")
			}
			if len(args) > 1 {
				if !processer.CopyFontsFromCache(files, args[1], nil) {
					ec++
				}
			}
		},
	}

	cmd.Flags().StringVarP(&f, "font-dir", "f", "", "Specify the ass fonts folder.")

	return cmd
}

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query <mkv-file|mkv-dir> [output-=file]",
		Short:   "Query folder for items",
		Aliases: []string{"q"},
		Args:    cobra.RangeArgs(1, 2),
		Long:    `Check a folder for items and generate a list of found items.`,
		Run: func(cmd *cobra.Command, args []string) {
			s := args[0]
			if i, _ := os.Stat(s); i.IsDir() {
				lines := processer.QueryFolder(s, nil)
				if len(lines) > 0 {
					data := []byte(strings.Join(lines, "\n"))
					f := "result.txt"
					if len(args) > 1 {
						f = args[1]
					}
					if os.WriteFile(f, data, os.ModePerm) != nil {
						color.Red("Failed to write the result file.")
						ec++
					}
				} else {
					color.Blue("No items found.")
				}
			} else {
				r, err := processer.CheckSubset(s, nil)
				if err {
					ec++
				} else {
					color.Blue("Need font subset: %v", !r)
				}
			}
		},
	}

	return cmd
}

func dumpCmd() *cobra.Command {
	n := false
	cmd := &cobra.Command{
		Use:     "dump <mkv-file|mkv-dir> [output-dir]",
		Short:   "Dump mkv file(s)",
		Aliases: []string{"d"},
		Args:    cobra.RangeArgs(1, 2),
		Long:    `Extract subtitles and fonts from mkv file or folder containing mkv files.`,
		Run: func(cmd *cobra.Command, args []string) {
			f := args[0]
			data := "data"
			if len(args) > 1 {
				data = args[1]
			}
			if i, _ := os.Stat(f); i.IsDir() {
				if !processer.DumpMKVs(f, data, !n, nil) {
					ec++
				}
			} else {
				if !processer.DumpMKV(f, data, !n, nil) {
					ec++
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&n, "no-subset", "n", false, "Don't subset")

	return cmd
}

func makeCmd() *cobra.Command {
	_l := ""
	_t := ""
	_mks := false
	_no := false
	cmd := &cobra.Command{
		Use:     "make <video-dir> [data-dir] [output-dir]",
		Short:   "Make mkv files",
		Aliases: []string{"m"},
		Args:    cobra.RangeArgs(1, 3),
		Long:    `Combine extracted subtitles and fonts back into mkv files.`,
		Run: func(cmd *cobra.Command, args []string) {
			data := "data"
			if len(args) > 1 {
				data = args[1]
			}
			dist := "dist"
			if len(args) > 2 {
				dist = args[2]
			}
			processer.MKS(_mks)
			processer.NOverwrite(_no)
			if !processer.MakeMKVs(args[0], data, dist, _l, _t, !nck, nil) {
				ec++
			}
		},
	}

	cmd.Flags().StringVarP(&_l, "subtitle-language", "l", "chi", "Specify the subtitle language.")
	cmd.Flags().StringVarP(&_t, "subtitle-title", "t", "", "Specify the subtitle title.")
	cmd.Flags().BoolVarP(&_mks, "enable-mks-output", "m", false, "Enable mks output.")
	cmd.Flags().BoolVarP(&_no, "Disable-overwrite", "n", false, "Disable file overwrite.")

	return cmd
}

func createCmd() *cobra.Command {
	_v := ""
	_s := ""
	_f := ""
	_o := ""
	_l := ""
	_t := ""
	_c := false
	_mks := false
	_no := false
	cmd := &cobra.Command{
		Use:     "create <input-dir>",
		Aliases: []string{"c"},
		Args:    cobra.ExactArgs(1),
		Short:   "Create mkv files",
		Long:    `Create mkv files from specified source folders, including videos, subtitles, and fonts.`,
		Run: func(cmd *cobra.Command, args []string) {
			p := args[0]
			v := path.Join(p, _v)
			s := path.Join(p, _s)
			f := path.Join(p, _f)
			o := path.Join(p, _o)
			processer.MKS(_mks)
			processer.NOverwrite(_no)
			if !processer.CreateMKVs(v, s, f, "", o, _l, _t, _c, nil) {
				ec++
			}
		},
	}

	cmd.Flags().StringVarP(&_v, "video-sub-dir", "v", "v", "Video sub dir.")
	cmd.Flags().StringVarP(&_s, "subtitle-sub-dir", "s", "s", "Subtitle sub dir.")
	cmd.Flags().StringVarP(&_f, "font-sub-dir", "f", "f", "Font sub dir.")
	cmd.Flags().StringVarP(&_o, "output-sub-dir", "o", "o", "Output sub dir.")

	cmd.Flags().StringVarP(&_l, "subtitle-language", "l", "chi", "Specify the subtitle language.")
	cmd.Flags().StringVarP(&_t, "subtitle-title", "t", "", "Specify the subtitle title.")
	cmd.Flags().BoolVarP(&_c, "clean", "c", false, "Clean original file subtitles and fonts.")
	cmd.Flags().BoolVarP(&_mks, "enable-mks-output", "m", false, "Enable mks output.")
	cmd.Flags().BoolVarP(&_no, "Disable-overwrite", "n", false, "Disable file overwrite.")

	return cmd
}

func subsetCmd() *cobra.Command {
	e := ""
	f := ""
	o := ""
	t := ""
	n := false
	b := false
	cmd := &cobra.Command{
		Use:     "subset <ass-file...|ass-dir>",
		Aliases: []string{"s"},
		Args:    cobra.MinimumNArgs(1),
		Short:   "Perform ass font subset",
		Long:    `Subset the fonts used in ass files and optionally create test videos with the subsetted fonts.`,
		Run: func(cmd *cobra.Command, args []string) {
			files := args
			if i, _ := os.Stat(files[0]); i.IsDir() {
				files, _ = findPath(files[0], `\.ass`)
			}
			if !processer.ASSFontSubset(files, f, o, !n, nil) {
				ec++
			} else if t != "" {
				d, _, _, _ := splitPath((files)[0])
				if o == "" {
					o = path.Join(d, "subsetted")
				} else if !n {
					o = path.Join(o, "subsetted")
				}
				files, _ = findPath(o, `\.ass$`)
				if len(files) > 0 {
					processer.CreateTestVideo(files, t, o, e, b, nil)
				}
			}
		},
	}
	cmd.Flags().StringVarP(&f, "font-dir", "f", "", "Specify the fonts folder.")
	cmd.Flags().StringVarP(&o, "output-dir", "o", "", "Specify the output folder.")
	cmd.Flags().BoolVarP(&n, "no-create-sub-dir", "n", false, `Do not output files to the new "subsetted" folder.`)
	cmd.Flags().StringVarP(&t, "test-video", "v", "", `Specify the source path to create a test video (enter "-" for a blank video).`)
	cmd.Flags().BoolVarP(&b, "video-burn-subtitle", "b", false, `Create a test video with burned-in subtitles.`)
	cmd.Flags().StringVarP(&e, "video-encoder", "e", "libx264", `Specify the encoder to use for creating the test video.`)

	return cmd
}

func cacheCmd() *cobra.Command {
	c := false
	cmd := &cobra.Command{
		Use:   "cache <font-dir>",
		Args:  cobra.ExactArgs(1),
		Short: "Create fonts cache",
		Long:  `Read font information from the specified directory and create a cache.`,
		Run: func(cmd *cobra.Command, args []string) {
			if c {
				_ = os.RemoveAll(cache)
			}
			s := args[0]
			p := path.Join(cache, path2MD5(s)+".cache")
			list := processer.CreateFontsCache(s, p, nil)
			el := len(list)
			if el > 0 {
				ec++
				color.Yellow("Error list:(%d)\n%s", el, strings.Join(list, "\n"))
			}
		},
	}

	cmd.Flags().BoolVarP(&c, "clean-old-cache", "c", false, "Clean old cache.")

	return cmd
}

func getLatestTag() {
	if resp, err := http.DefaultClient.Get("https://api.github.com/repos/MkvAutoSubset/MkvAutoSubset/releases/latest"); err == nil {
		if data, err := io.ReadAll(resp.Body); err == nil {
			reg, _ := regexp.Compile(`"tag_name":"([^"]+)"`)
			arr := reg.FindStringSubmatch(string(data))
			if len(arr) > 1 {
				latestTag = arr[1]
			}
		}
	}
}
