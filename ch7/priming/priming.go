// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// priming illustrates _weight-based priming_, that is, how small
// weight changes caused by the standard slow cortical learning rate
// can produce significant behavioral priming, causing the network
// to favor one output pattern over another.
package main

//go:generate core generate -add-types

import (
	"embed"
	"fmt"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/randx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/plot/plotcore"
	"cogentcore.org/core/tensor/table"
	"cogentcore.org/core/tree"
	"github.com/emer/emergent/v2/econfig"
	"github.com/emer/emergent/v2/egui"
	"github.com/emer/emergent/v2/elog"
	"github.com/emer/emergent/v2/emer"
	"github.com/emer/emergent/v2/env"
	"github.com/emer/emergent/v2/estats"
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/looper"
	"github.com/emer/emergent/v2/netview"
	"github.com/emer/emergent/v2/params"
	"github.com/emer/emergent/v2/paths"
	"github.com/emer/leabra/v2/leabra"
)

//go:embed twout_all.tsv twout_a.tsv twout_b.tsv trained.wts
var content embed.FS

func main() {
	sim := &Sim{}
	sim.New()
	sim.ConfigAll()
	sim.RunGUI()
}

// EnvTypes are the types of train / test environments.
type EnvTypes int32 //enums:enum

const (
	// TrainB sets train env to TrainB pats
	TrainB EnvTypes = iota

	// TrainA sets train env to TrainA pats
	TrainA

	// TrainAll sets train to TrainAll pats
	TrainAll

	// TestA sets testing to TrainA pats, for wt priming
	TestA

	// TestB sets testing to TrainB pats
	TestB

	// TestAll sets testing to TrainAll pats, for act priming
	TestAll
)

// ParamSets is the default set of parameters.
// Base is always applied, and others can be optionally
// selected to apply on top of that.
var ParamSets = params.Sets{
	"Base": {
		{Sel: "Path", Desc: "no extra learning factors",
			Params: params.Params{
				"Path.Learn.Norm.On":     "false",
				"Path.Learn.Momentum.On": "false",
				"Path.Learn.WtBal.On":    "false",
				"Path.Learn.Lrate":       "0.04",
			}},
		{Sel: "Layer", Desc: "less inhib for smaller in / out layers",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi":    "1.5",
				"Layer.Inhib.ActAvg.Init": "0.25",
				"Layer.Act.Gbar.L":        "0.1",
				"Layer.Act.Init.Decay":    "1",
			}},
		{Sel: "#Hidden", Desc: "slightly less inhib",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "1.6",
			}},
		{Sel: ".BackPath", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates",
			Params: params.Params{
				"Path.WtScale.Rel": "0.1",
			}},
	},
}

// Config has config parameters related to running the sim
type Config struct {
	// total number of runs to do when running Train
	NRuns int `default:"1" min:"1"`

	// total number of epochs per run
	NEpochs int `default:"100"`

	// how often to run through all the test patterns, in terms of training epochs.
	// can use 0 or -1 for no testing.
	TestInterval int `default:"1"`
}

// Sim encapsulates the entire simulation model, and we define all the
// functionality as methods on this struct.  This structure keeps all relevant
// state information organized and available without having to pass everything around
// as arguments to methods, and provides the core GUI interface (note the view tags
// for the fields which provide hints to how things should be displayed).
type Sim struct {

	// Lrate is the learning rate; .04 is default 'cortical' learning rate.
	// Try lower levels to see how low you can go and still get priming.
	Lrate float32 `def:"0.04"`

	// Decay is the proportion of activation decay between trials.
	Decay float32 `def:"1"`

	// EnvType is the environment type; Use the Env button (SetEnv) to set.
	EnvType EnvTypes `read-only:"+"`

	// Config contains misc configuration parameters for running the sim
	Config Config `new-window:"+" display:"no-inline"`

	// the network -- click to view / edit parameters for layers, paths, etc
	Net *leabra.Network `new-window:"+" display:"no-inline"`

	// network parameter management
	Params emer.NetParams `display:"add-fields"`

	// All training patterns
	TrainAll *table.Table `new-window:"+" display:"no-inline"`

	// A training patterns
	TrainA *table.Table `new-window:"+" display:"no-inline"`

	// B training patterns
	TrainB *table.Table `new-window:"+" display:"no-inline"`

	// contains looper control loops for running sim
	Loops *looper.Manager `new-window:"+" display:"no-inline"`

	// contains computed statistic values
	Stats estats.Stats `new-window:"+"`

	// Contains all the logs and information about the logs.'
	Logs elog.Logs `new-window:"+"`

	// Environments
	Envs env.Envs `new-window:"+" display:"no-inline"`

	// leabra timing parameters and state
	Context leabra.Context `new-window:"+"`

	// netview update parameters
	ViewUpdate netview.ViewUpdate `display:"add-fields"`

	// manages all the gui elements
	GUI egui.GUI `display:"-"`

	// a list of random seeds to use for each run
	RandSeeds randx.Seeds `display:"-"`
}

// New creates new blank elements and initializes defaults
func (ss *Sim) New() {
	ss.Defaults()
	econfig.Config(&ss.Config, "config.toml")
	ss.Net = leabra.NewNetwork("Priming")
	ss.Params.Config(ParamSets, "", "", ss.Net)
	ss.Stats.Init()
	ss.Stats.SetInt("Expt", 0)
	ss.TrainAll = &table.Table{}
	ss.TrainA = &table.Table{}
	ss.TrainB = &table.Table{}
	ss.RandSeeds.Init(100) // max 100 runs
	ss.InitRandSeed(0)
	ss.Context.Defaults()
}

func (ss *Sim) Defaults() {
	ss.Lrate = 0.04
	ss.EnvType = TrainAll
	ss.Decay = 1
}

//////////////////////////////////////////////////////////////////////////////
// 		Configs

// ConfigAll configures all the elements using the standard functions
func (ss *Sim) ConfigAll() {
	ss.OpenPatterns()
	ss.ConfigEnv()
	ss.ConfigNet(ss.Net)
	ss.ConfigLogs()
	ss.ConfigLoops()
}

// OpenPatAsset opens pattern file from embedded assets
func (ss *Sim) OpenPatAsset(dt *table.Table, fnm, name, desc string) error {
	dt.SetMetaData("name", name)
	dt.SetMetaData("desc", desc)
	err := dt.OpenFS(content, fnm, table.Tab)
	if errors.Log(err) == nil {
		for i := 1; i < dt.NumColumns(); i++ {
			dt.Columns[i].SetMetaData("grid-fill", "0.9")
		}
	}
	return err
}

func (ss *Sim) OpenPatterns() {
	ss.OpenPatAsset(ss.TrainAll, "twout_all.tsv", "TrainAll", "All Training Patterns")
	ss.OpenPatAsset(ss.TrainA, "twout_a.tsv", "TrainA", "A Training Patterns")
	ss.OpenPatAsset(ss.TrainB, "twout_b.tsv", "TrainB", "B Training Patterns")
}

func (ss *Sim) ConfigEnv() {
	// Can be called multiple times -- don't re-create
	var trn, tst *env.FixedTable
	if len(ss.Envs) == 0 {
		trn = &env.FixedTable{}
		tst = &env.FixedTable{}
	} else {
		trn = ss.Envs.ByMode(etime.Train).(*env.FixedTable)
		tst = ss.Envs.ByMode(etime.Test).(*env.FixedTable)
	}

	// note: names must be standard here!
	trn.Name = etime.Train.String()
	trn.Config(table.NewIndexView(ss.TrainAll))

	tst.Name = etime.Test.String()
	tst.Config(table.NewIndexView(ss.TrainA))
	tst.Sequential = true

	trn.Init(0)
	tst.Init(0)

	// note: names must be in place when adding
	ss.Envs.Add(trn, tst)
}

func (ss *Sim) ConfigNet(net *leabra.Network) {
	net.SetRandSeed(ss.RandSeeds[0]) // init new separate random seed, using run = 0

	inp := net.AddLayer2D("Input", 5, 5, leabra.InputLayer)
	hid := net.AddLayer2D("Hidden", 6, 6, leabra.SuperLayer)
	out := net.AddLayer2D("Output", 5, 5, leabra.TargetLayer)

	full := paths.NewFull()

	net.ConnectLayers(inp, hid, full, leabra.ForwardPath)
	net.BidirConnectLayers(hid, out, full)

	net.Build()
	net.Defaults()
	ss.ApplyParams()
	net.InitWeights()
}

func (ss *Sim) ApplyParams() {
	if ss.Loops != nil {
		trn := ss.Loops.Stacks[etime.Train]
		trn.Loops[etime.Run].Counter.Max = ss.Config.NRuns
		trn.Loops[etime.Epoch].Counter.Max = ss.Config.NEpochs
	}

	spo := errors.Log1(errors.Log1(ss.Params.Params.SheetByName("Base")).SelByName("Path"))
	spo.Params.SetByName("Path.Learn.Lrate", fmt.Sprintf("%g", ss.Lrate))

	spo = errors.Log1(errors.Log1(ss.Params.Params.SheetByName("Base")).SelByName("Layer"))
	spo.Params.SetByName("Layer.Act.Init.Decay", fmt.Sprintf("%g", ss.Decay))

	ss.Params.SetAll()
}

////////////////////////////////////////////////////////////////////////////////
// 	    Init, utils

// Init restarts the run, and initializes everything, including network weights
// and resets the epoch log table
func (ss *Sim) Init() {
	ss.Stats.SetString("RunName", ss.Params.RunName(0)) // in case user interactively changes tag
	ss.Loops.ResetCounters()
	ss.InitRandSeed(0)
	ss.ConfigEnv() // re-config env just in case a different set of patterns was
	ss.GUI.StopNow = false
	ss.ApplyParams()
	ss.NewRun()
	ss.ViewUpdate.RecordSyns()
	ss.ViewUpdate.Update()
}

// InitRandSeed initializes the random seed based on current training run number
func (ss *Sim) InitRandSeed(run int) {
	ss.RandSeeds.Set(run)
	ss.RandSeeds.Set(run, &ss.Net.Rand)
}

// ConfigLoops configures the control loops: Training, Testing
func (ss *Sim) ConfigLoops() {
	man := looper.NewManager()

	trls := ss.TrainAll.Rows
	ttrls := ss.TrainA.Rows

	man.AddStack(etime.Train).
		AddTime(etime.Run, ss.Config.NRuns).
		AddTime(etime.Epoch, ss.Config.NEpochs).
		AddTime(etime.Trial, trls).
		AddTime(etime.Cycle, 100)

	man.AddStack(etime.Test).
		AddTime(etime.Epoch, 1).
		AddTime(etime.Trial, ttrls).
		AddTime(etime.Cycle, 100)

	leabra.LooperStdPhases(man, &ss.Context, ss.Net, 75, 99)                // plus phase timing
	leabra.LooperSimCycleAndLearn(man, ss.Net, &ss.Context, &ss.ViewUpdate) // std algo code

	for m, _ := range man.Stacks {
		stack := man.Stacks[m]
		stack.Loops[etime.Trial].OnStart.Add("ApplyInputs", func() {
			ss.ApplyInputs()
		})
	}

	man.GetLoop(etime.Train, etime.Run).OnStart.Add("NewRun", ss.NewRun)

	man.GetLoop(etime.Train, etime.Run).OnEnd.Add("RunDone", func() {
		if ss.Stats.Int("Run") >= ss.Config.NRuns-1 {
			ss.RunStats()
			expt := ss.Stats.Int("Expt")
			ss.Stats.SetInt("Expt", expt+1)
		}
	})

	// Add Testing
	trainEpoch := man.GetLoop(etime.Train, etime.Epoch)
	trainEpoch.OnStart.Add("TestAtInterval", func() {
		if (ss.Config.TestInterval > 0) && ((trainEpoch.Counter.Cur+1)%ss.Config.TestInterval == 0) {
			// Note the +1 so that it doesn't occur at the 0th timestep.
			ss.TestAll()
		}
	})

	/////////////////////////////////////////////
	// Logging

	man.GetLoop(etime.Test, etime.Epoch).OnEnd.Add("LogTestErrors", func() {
		leabra.LogTestErrors(&ss.Logs)
	})
	man.AddOnEndToAll("Log", ss.Log)
	leabra.LooperResetLogBelow(man, &ss.Logs)
	man.GetLoop(etime.Train, etime.Run).OnEnd.Add("RunStats", func() {
		ss.Logs.RunStats("PctCor", "FirstZero", "LastZero")
	})

	////////////////////////////////////////////
	// GUI

	leabra.LooperUpdateNetView(man, &ss.ViewUpdate, ss.Net, ss.NetViewCounters)
	leabra.LooperUpdatePlots(man, &ss.GUI)

	ss.Loops = man
}

// ApplyInputs applies input patterns from given environment.
// It is good practice to have this be a separate method with appropriate
// args so that it can be used for various different contexts
// (training, testing, etc).
func (ss *Sim) ApplyInputs() {
	ctx := &ss.Context
	net := ss.Net
	ev := ss.Envs.ByMode(ctx.Mode).(*env.FixedTable)
	ev.Step()

	lays := net.LayersByType(leabra.InputLayer, leabra.TargetLayer)
	net.InitExt()
	ss.Stats.SetString("TrialName", ev.TrialName.Cur)
	for _, lnm := range lays {
		ly := ss.Net.LayerByName(lnm)
		pats := ev.State(ly.Name)
		if pats != nil {
			ly.ApplyExt(pats)
		}
	}
}

// NewRun intializes a new run of the model, using the TrainEnv.Run counter
// for the new run value
func (ss *Sim) NewRun() {
	ctx := &ss.Context
	ss.InitRandSeed(ss.Loops.GetLoop(etime.Train, etime.Run).Counter.Cur)
	ss.Envs.ByMode(etime.Train).Init(0)
	ss.Envs.ByMode(etime.Test).Init(0)
	ctx.Reset()
	ctx.Mode = etime.Train
	ss.Net.InitWeights()
	ss.InitStats()
	ss.StatCounters()
	ss.Logs.ResetLog(etime.Train, etime.Epoch)
	ss.Logs.ResetLog(etime.Test, etime.Epoch)
}

// SetEnv select which set of patterns to train on: AB or AC
func (ss *Sim) SetEnv(envType EnvTypes) { //types:add
	trn := ss.Envs.ByMode(etime.Train).(*env.FixedTable)
	tst := ss.Envs.ByMode(etime.Test).(*env.FixedTable)
	ss.EnvType = envType
	switch envType {
	case TrainA:
		trn.Table = table.NewIndexView(ss.TrainA)
		trn.Init(0)
	case TrainB:
		trn.Table = table.NewIndexView(ss.TrainB)
		trn.Init(0)
	case TrainAll:
		trn.Table = table.NewIndexView(ss.TrainAll)
		trn.Init(0)
	case TestA:
		tst.Table = table.NewIndexView(ss.TrainA)
		tst.Init(0)
	case TestB:
		tst.Table = table.NewIndexView(ss.TrainB)
		tst.Init(0)
	case TestAll:
		tst.Table = table.NewIndexView(ss.TrainAll)
		tst.Init(0)
	}
}

// TestAll runs through the full set of testing items
func (ss *Sim) TestAll() {
	ss.Envs.ByMode(etime.Test).Init(0)
	ss.Loops.ResetAndRun(etime.Test)
	ss.Loops.Mode = etime.Train // Important to reset Mode back to Train because this is called from within the Train Run.
}

////////////////////////////////////////////////////////////////////////
// 		Stats

// InitStats initializes all the statistics.
// called at start of new run
func (ss *Sim) InitStats() {
	ss.Stats.SetFloat("SSE", 0.0)
	ss.Stats.SetString("TrialName", "")
	ss.Stats.SetString("Closest", "")
	ss.Stats.SetFloat("Correl", 0.0)
	ss.Stats.SetFloat("TrlErr", 0.0)
	ss.Stats.SetFloat("IsA", 0.0)
	ss.Stats.SetFloat("IsB", 0.0)
	ss.Logs.InitErrStats() // inits TrlErr, FirstZero, LastZero, NZero
}

// StatCounters saves current counters to Stats, so they are available for logging etc
// Also saves a string rep of them for ViewUpdate.Text
func (ss *Sim) StatCounters() {
	ctx := &ss.Context
	mode := ctx.Mode
	ss.Loops.Stacks[mode].CountersToStats(&ss.Stats)
	// always use training epoch..
	trnEpc := ss.Loops.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
	ss.Stats.SetInt("Epoch", trnEpc)
	trl := ss.Stats.Int("Trial")
	ss.Stats.SetInt("Trial", trl)
	ss.Stats.SetInt("Cycle", int(ctx.Cycle))
}

func (ss *Sim) NetViewCounters(tm etime.Times) {
	if ss.ViewUpdate.View == nil {
		return
	}
	if tm == etime.Trial {
		ss.TrialStats() // get trial stats for current di
	}
	ss.StatCounters()
	ss.ViewUpdate.Text = ss.Stats.Print([]string{"Run", "Epoch", "Trial", "TrialName", "Cycle", "SSE", "TrlErr", "IsA", "IsB"})
}

// TrialStats computes the trial-level statistics.
// Aggregation is done directly from log data.
func (ss *Sim) TrialStats() {
	out := ss.Net.LayerByName("Output")
	sse, avgsse := out.MSE(0.5) // 0.5 = per-unit tolerance -- right side of .5
	ss.Stats.SetFloat("SSE", sse)
	ss.Stats.SetFloat("AvgSSE", avgsse)

	_, cor, cnm := ss.Stats.ClosestPat(ss.Net, "Output", "ActM", 0, ss.TrainAll, "Output", "Name")
	ss.Stats.SetString("Closest", cnm)
	ss.Stats.SetFloat32("Correl", cor)

	tnm := ss.Stats.String("TrialName")
	cnmsp := strings.Split(cnm, "_")
	tnmsp := strings.Split(tnm, "_")
	if cnmsp[0] == tnmsp[0] {
		ss.Stats.SetFloat("TrlErr", 0)
	} else {
		ss.Stats.SetFloat("TrlErr", 1)
	}
	if cnmsp[1] == "a" {
		ss.Stats.SetFloat("IsA", 1)
	} else {
		ss.Stats.SetFloat("IsA", 0)
	}
	ss.Stats.SetFloat("IsB", 1-ss.Stats.Float("IsA"))
}

func (ss *Sim) TestStats() {
	trl := ss.Logs.Table(etime.Test, etime.Trial)
	if trl.Rows == 0 {
		return
	}
	// trix := table.NewIndexView(trl)
	// spl := split.GroupBy(trix, "GroupName")
	// split.AggColumn(spl, "Err", stats.Mean)
	// tsts := spl.AggsToTable(table.ColumnNameOnly)
	// ss.Logs.MiscTables["TestEpoch"] = tsts
	// ss.Stats.SetFloat("ABErr", tsts.Columns[1].Float1D(0))
	// ss.Stats.SetFloat("ACErr", tsts.Columns[1].Float1D(1))
}

func (ss *Sim) RunStats() {
	// dt := ss.Logs.Table(etime.Train, etime.Run)
	// runix := table.NewIndexView(dt)
	// spl := split.GroupBy(runix, "Expt")
	// split.DescColumn(spl, "ABErr")
	// st := spl.AggsToTableCopy(table.AddAggName)
	// ss.Logs.MiscTables["RunStats"] = st
	// plt := ss.GUI.Plots[etime.ScopeKey("RunStats")]
	//
	// st.SetMetaData("XAxis", "RunName")
	//
	// st.SetMetaData("Points", "true")
	//
	// st.SetMetaData("ABErr:Mean:On", "+")
	// st.SetMetaData("ABErr:Mean:FixMin", "true")
	// st.SetMetaData("ABErr:Mean:FixMax", "true")
	// st.SetMetaData("ABErr:Mean:Min", "0")
	// st.SetMetaData("ABErr:Mean:Max", "1")
	// st.SetMetaData("ABErr:Min:On", "+")
	// st.SetMetaData("ABErr:Count:On", "-")
	//
	// plt.SetTable(st)
	// plt.GoUpdatePlot()
}

//////////////////////////////////////////////////////////////////////
// 		Logging

func (ss *Sim) ConfigLogs() {
	ss.Stats.SetString("RunName", ss.Params.RunName(0)) // used for naming logs, stats, etc

	ss.Logs.AddCounterItems(etime.Run, etime.Epoch, etime.Trial, etime.Cycle)
	ss.Logs.AddStatIntNoAggItem(etime.AllModes, etime.AllTimes, "Expt")
	ss.Logs.AddStatStringItem(etime.AllModes, etime.AllTimes, "RunName")
	ss.Logs.AddStatStringItem(etime.AllModes, etime.Trial, "TrialName")
	ss.Logs.AddStatStringItem(etime.AllModes, etime.Trial, "Closest")

	ss.Logs.AddStatAggItem("SSE", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("AvgSSE", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("Correl", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("IsA", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("IsB", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddErrStatAggItems("TrlErr", etime.Run, etime.Epoch, etime.Trial)

	ss.Logs.AddPerTrlMSec("PerTrlMSec", etime.Run, etime.Epoch, etime.Trial)

	ss.Logs.AddLayerTensorItems(ss.Net, "ActM", etime.Test, etime.Trial, "InputLayer", "SuperLayer", "TargetLayer")
	ss.Logs.AddLayerTensorItems(ss.Net, "Targ", etime.Test, etime.Trial, "TargetLayer")

	ss.Logs.PlotItems("PctErr", "Correl")

	ss.Logs.CreateTables()
	ss.Logs.SetContext(&ss.Stats, ss.Net)
	// don't plot certain combinations we don't use
	ss.Logs.NoPlot(etime.Train, etime.Cycle)
	ss.Logs.NoPlot(etime.Test, etime.Cycle)
	// ss.Logs.NoPlot(etime.Test, etime.Trial)
	ss.Logs.NoPlot(etime.Test, etime.Run)
	ss.Logs.SetMeta(etime.Train, etime.Run, "LegendCol", "RunName")

	ss.Logs.SetMeta(etime.Test, etime.Trial, "Closest:On", "+")
	ss.Logs.SetMeta(etime.Test, etime.Trial, "Correl:On", "-")
	ss.Logs.SetMeta(etime.Test, etime.Trial, "IsA:On", "+")
}

// Log is the main logging function, handles special things for different scopes
func (ss *Sim) Log(mode etime.Modes, time etime.Times) {
	ctx := &ss.Context
	if mode != etime.Analyze {
		ctx.Mode = mode // Also set specifically in a Loop callback.
	}
	dt := ss.Logs.Table(mode, time)
	if dt == nil {
		return
	}
	row := dt.Rows

	switch {
	case time == etime.Cycle:
		return
	case time == etime.Trial:
		ss.TrialStats()
		ss.StatCounters()
	case time == etime.Epoch && mode == etime.Test:
		ss.TestStats()
	}

	ss.Logs.LogRow(mode, time, row) // also logs to file, etc

	// if mode == etime.Test {
	// 	ss.GUI.UpdateTableView(etime.Test, etime.Trial)
	// }
}

//////////////////////////////////////////////////////////////////////
// 		GUI

// ConfigGUI configures the Cogent Core GUI interface for this simulation.
func (ss *Sim) ConfigGUI() {
	title := "Priming"
	ss.GUI.MakeBody(ss, "priming", title, `This simulation explores the neural basis of *priming* -- the often surprisingly strong impact of residual traces from prior experience, which can be either *weight-based* (small changes in synapses) or *activation-based* (residual neural activity).  In the first part, we see how small weight changes caused by the standard slow cortical learning rate can produce significant behavioral priming, causing the network to favor one output pattern over another.  Likewise, residual activation can bias subsequent processing, but this is short-lived and transient compared to the long-lasting effects of weight-based priming. See <a href="https://github.com/CompCogNeuro/sims/blob/main/ch7/priming/README.md">README.md on GitHub</a>.</p>`)
	ss.GUI.CycleUpdateInterval = 10

	nv := ss.GUI.AddNetView("Network")
	nv.Options.MaxRecs = 300
	nv.SetNet(ss.Net)
	nv.Options.PathWidth = 0.003
	ss.ViewUpdate.Config(nv, etime.GammaCycle, etime.GammaCycle)
	ss.GUI.ViewUpdate = &ss.ViewUpdate
	nv.Current()

	ss.GUI.AddPlots(title, &ss.Logs)

	stnm := "RunStats"
	dt := ss.Logs.MiscTable(stnm)
	bcp, _ := ss.GUI.Tabs.NewTab(stnm + " Plot")
	plt := plotcore.NewSubPlot(bcp)
	ss.GUI.Plots[etime.ScopeKey(stnm)] = plt
	plt.Options.Title = "Run Stats"
	plt.Options.XAxis = "RunName"
	plt.SetTable(dt)

	// ss.GUI.AddTableView(&ss.Logs, etime.Test, etime.Trial)

	ss.GUI.FinalizeGUI(false)
}

func (ss *Sim) MakeToolbar(p *tree.Plan) {
	ss.GUI.AddToolbarItem(p, egui.ToolbarItem{Label: "Init", Icon: icons.Update,
		Tooltip: "Initialize everything including network weights, and start over.  Also applies current params.",
		Active:  egui.ActiveStopped,
		Func: func() {
			ss.Init()
			ss.GUI.UpdateWindow()
		},
	})

	ss.GUI.AddLooperCtrl(p, ss.Loops, []etime.Modes{etime.Train, etime.Test})

	ss.GUI.AddToolbarItem(p, egui.ToolbarItem{Label: "Test Init", Icon: icons.Update,
		Tooltip: "Initialize testing to start over -- if Test Step doesn't work, then do this.",
		Active:  egui.ActiveStopped,
		Func: func() {
			ss.Loops.ResetCountersByMode(etime.Test)
		},
	})

	ss.GUI.AddToolbarItem(p, egui.ToolbarItem{Label: "Set inputs",
		Icon:    icons.Settings,
		Tooltip: "Set the input patterns to use for training and testing",
		Active:  egui.ActiveAlways,
		Func: func() {
			core.CallFunc(ss.GUI.Body, ss.SetEnv)
		},
	})

	ss.GUI.AddToolbarItem(p, egui.ToolbarItem{Label: "Open Trained Wts",
		Icon:    icons.Open,
		Tooltip: "Open trained weights, trained on the Train All patterns",
		Active:  egui.ActiveAlways,
		Func: func() {
			ss.Net.OpenWeightsFS(content, "trained.wts")
		},
	})

	////////////////////////////////////////////////
	tree.Add(p, func(w *core.Separator) {})
	ss.GUI.AddToolbarItem(p, egui.ToolbarItem{Label: "New Seed",
		Icon:    icons.Add,
		Tooltip: "Generate a new initial random seed to get different results.  By default, Init re-establishes the same initial seed every time.",
		Active:  egui.ActiveAlways,
		Func: func() {
			ss.RandSeeds.NewSeeds()
		},
	})
	ss.GUI.AddToolbarItem(p, egui.ToolbarItem{Label: "README",
		Icon:    icons.FileMarkdown,
		Tooltip: "Opens your browser on the README file that contains instructions for how to run this model.",
		Active:  egui.ActiveAlways,
		Func: func() {
			core.TheApp.OpenURL("https://github.com/CompCogNeuro/sims/blob/main/ch4/family_trees/README.md")
		},
	})
}

func (ss *Sim) RunGUI() {
	ss.Init()
	ss.ConfigGUI()
	ss.GUI.Body.RunMainWindow()
}
