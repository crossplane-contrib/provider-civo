/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"path/filepath"

	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	civoinstance "github.com/crossplane-contrib/provider-civo/internal/controller/civoinstance"

	log "github.com/sirupsen/logrus"

	"gopkg.in/alecthomas/kingpin.v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/crossplane-contrib/provider-civo/apis"
	civokubernetes "github.com/crossplane-contrib/provider-civo/internal/controller/civokubernetes"
	"github.com/crossplane-contrib/provider-civo/internal/controller/civovolume"
	civoprovider "github.com/crossplane-contrib/provider-civo/internal/controller/provider"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	var (
		app              = kingpin.New(filepath.Base(os.Args[0]), "Template support for Crossplane.").DefaultEnvars()
		debug            = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		syncPeriod       = app.Flag("sync", "Controller manager sync period such as 300ms, 1.5h, or 2h45m").Short('s').Default("1h").Duration()
		leaderElection   = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool()
		maxReconcileRate = app.Flag("max-reconcile-rate", "The number of concurrent reconciliations that may be running at one time.").Default("100").Int()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-template"))
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	if *debug {
		// The controller-runtime runs with a no-op logger by default. It is
		// *very* verbose even at info level, so we only provide it a real
		// logger when we're running in debug mode.
		ctrl.SetLogger(zl)
	}

	log.Debug("Starting", "sync-period", syncPeriod.String())

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		LeaderElection:   *leaderElection,
		LeaderElectionID: "crossplane-leader-election-provider-template",
		Cache: cache.Options{
			SyncPeriod: syncPeriod,
		},
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	rateLimiter := ratelimiter.NewGlobal(*maxReconcileRate)

	kingpin.FatalIfError(apis.AddToScheme(mgr.GetScheme()), "Cannot add Template APIs to scheme")
	kingpin.FatalIfError(civokubernetes.Setup(mgr, log, *rateLimiter), "Cannot setup Civo K3 Cluster controllers")
	kingpin.FatalIfError(civoinstance.Setup(mgr, log, *rateLimiter), "Cannot setup Civo Instance controllers")
	kingpin.FatalIfError(civovolume.Setup(mgr, log, *rateLimiter), "Cannot setup Civo volume controllers")
	kingpin.FatalIfError(civoprovider.Setup(mgr, log, *rateLimiter), "Cannot setup Provider controllers")
	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}
