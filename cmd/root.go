package cmd

import (
    "bufio"
    "fmt"
    "github.com/j3ssie/osmedeus/database"
    "os"
    "strings"

    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/spf13/cobra"
    "gorm.io/gorm"
)

var options = libs.Options{}

// DB database variables
var DB *gorm.DB

var RootCmd = &cobra.Command{
    Use:   fmt.Sprintf("%s", libs.BINARY),
    Short: fmt.Sprintf("%s - %s", libs.BINARY, libs.DESC),
    Long:  core.Banner(),
}

// Execute main function
func Execute() {
    if err := RootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func init() {
    RootCmd.PersistentFlags().StringVar(&options.Env.RootFolder, "rootFolder", fmt.Sprintf("~/.%s/", libs.BINARY), "Root Folder to store Result")
    RootCmd.PersistentFlags().StringVar(&options.Env.BaseFolder, "baseFolder", fmt.Sprintf("~/%s-base/", libs.BINARY), "Base Folder which is store data, binaries and workflows")
    RootCmd.PersistentFlags().StringVar(&options.ConfigFile, "configFile", fmt.Sprintf("~/.%s/config.yaml", libs.BINARY), "Config File")
    RootCmd.PersistentFlags().StringVar(&options.LogFile, "log", "", "Log File")
    RootCmd.PersistentFlags().IntVarP(&options.Concurrency, "concurrency", "c", 1, "Concurrency level (recommend to keep it as 1 on machine has RAM smaller than 2GB)")

    // parse target as global flag
    RootCmd.PersistentFlags().StringSliceVarP(&options.Scan.Inputs, "target", "t", []string{}, "Target to running")
    RootCmd.PersistentFlags().StringVarP(&options.Scan.InputList, "targets", "T", "", "List of target as a file")

    // cloud flags
    RootCmd.PersistentFlags().BoolVar(&options.Cloud.EnableChunk, "chunk", false, "Enable chunk mode")
    RootCmd.PersistentFlags().IntVarP(&options.Cloud.NumberOfParts, "chunk-parts", "P", 0, "Number of chunks file to split (default: equal with concurrency)")
    RootCmd.PersistentFlags().StringVar(&options.Cloud.ChunkInputs, "chunkFolder", "/tmp/chunk-inputs/", "Temp Folder to store chunk inputs")
    RootCmd.PersistentFlags().StringVar(&options.Timeout, "timeout", "", "Global timeout for each step (e.g: 60s, 30m, 2h)")
    RootCmd.PersistentFlags().StringVar(&options.Cloud.Size, "size", "", "Override Size of cloud instance (default will get from 'cloud/config.yaml')")
    RootCmd.PersistentFlags().StringVar(&options.Cloud.Region, "region", "", "Override Region of cloud instance (default will get from 'cloud/config.yaml')")
    RootCmd.PersistentFlags().StringVar(&options.Cloud.Token, "token", "", "Override token of cloud instance (default will get from 'cloud/config.yaml')")
    RootCmd.PersistentFlags().StringVar(&options.Cloud.Provider, "provider", "", "Override provider of cloud instance (default will get from 'cloud/config.yaml')")
    RootCmd.PersistentFlags().BoolVar(&options.Cloud.ReBuildBaseImage, "rebuild", false, "Forced to rebuild the images event though the version didn't change")

    // mics option
    RootCmd.PersistentFlags().StringVarP(&options.ScanID, "sid", "s", "", "Scan ID to continue the scan without create new scan record")
    RootCmd.PersistentFlags().BoolVarP(&options.Resume, "resume", "R", false, "Enable Resume")
    RootCmd.PersistentFlags().BoolVar(&options.Debug, "debug", false, "Enable Debug output")
    RootCmd.PersistentFlags().BoolVarP(&options.Quite, "quite", "q", false, "Show only essential information")
    RootCmd.PersistentFlags().BoolVar(&options.WildCardCheck, "ww", false, "Check for wildcard target")
    RootCmd.PersistentFlags().BoolVar(&options.DisableValidateInput, "nv", false, "Disable Validate Input")
    RootCmd.PersistentFlags().BoolVar(&options.Update.NoUpdate, "nu", false, "Disable Update options")
    RootCmd.PersistentFlags().BoolVarP(&options.Verbose, "verbose", "V", false, "Show stat info too")
    RootCmd.PersistentFlags().BoolVarP(&options.EnableFormatInput, "format-input", "J", false, "Enable special input format")

    // disable options
    RootCmd.PersistentFlags().BoolVar(&options.NoNoti, "nn", false, "No notification")
    RootCmd.PersistentFlags().BoolVar(&options.NoBanner, "nb", false, "No banner")
    RootCmd.PersistentFlags().BoolVarP(&options.NoDB, "no-db", "D", false, "No store DB record")
    RootCmd.PersistentFlags().BoolVarP(&options.NoGit, "no-git", "N", false, "No git storage")
    RootCmd.PersistentFlags().BoolVarP(&options.NoClean, "no-clean", "C", false, "No clean junk output")
    RootCmd.PersistentFlags().StringSliceVarP(&options.Exclude, "exclude", "x", []string{}, "Exclude module name (Multiple -x flags are accepted)")
    RootCmd.PersistentFlags().BoolVarP(&options.CustomGit, "git", "g", false, "Use custom Git repo")

    // sync options
    RootCmd.PersistentFlags().BoolVar(&options.EnableDeStorage, "des", false, "Enable Dedicated Storages")
    RootCmd.PersistentFlags().BoolVar(&options.GitSync, "sync", false, "Enable Sync Check before doing git push")
    RootCmd.PersistentFlags().IntVar(&options.SyncTimes, "sync-timee", 15, "Number of times to check before force push")
    RootCmd.PersistentFlags().IntVar(&options.PollingTime, "poll-timee", 100, "Number of seconds to sleep before do next sync check")
    RootCmd.PersistentFlags().BoolVar(&options.NoCdn, "no-cdn", false, "Disable CDN feature")
    RootCmd.PersistentFlags().BoolVarP(&options.EnableBackup, "backup", "b", false, "Enable Backup after done")
    RootCmd.PersistentFlags().BoolVar(&options.JsonOutput, "json", false, "Output as JSON")

    // update options
    RootCmd.PersistentFlags().BoolVar(&options.Update.IsUpdateBin, "bin", false, "Update binaries too")
    RootCmd.PersistentFlags().BoolVar(&options.Update.EnableUpdate, "update", false, "Enable auto update")
    RootCmd.PersistentFlags().StringVar(&options.Update.UpdateFolder, "update-folder", "/tmp/osm-update", "Folder to clone the update folder")

    RootCmd.SetHelpFunc(RootHelp)
    cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    if options.JsonOutput {
        options.Quite = true
    }

    /* Really Start the program */
    utils.InitLog(&options)
    core.InitConfig(&options)

    // parse inputs
    if options.Scan.InputList != "" {
        if utils.FileExists(options.Scan.InputList) {
            options.Scan.Inputs = append(options.Scan.Inputs, utils.ReadingFileUnique(options.Scan.InputList)...)
        }
    }

    // detect if anything came from stdin
    stat, _ := os.Stdin.Stat()
    if (stat.Mode() & os.ModeCharDevice) == 0 {
        sc := bufio.NewScanner(os.Stdin)
        for sc.Scan() {
            target := strings.TrimSpace(sc.Text())
            if err := sc.Err(); err == nil && target != "" {
                options.Scan.Inputs = append(options.Scan.Inputs, target)
            }
        }
    }

    if options.Update.EnableUpdate && options.Update.NoUpdate == false {
        core.Update(options)
    }
}

// DBInit init database connection
func DBInit() {
    var err error
    DB, err = database.InitDB(options)
    if err != nil {
        // simple retry
        DB, err = database.InitDB(options)
        if err != nil {
            fmt.Printf("[panic] Can't connect to DB at %v\n", options.Server.DBPath)
            os.Exit(-1)
        }
    }
}