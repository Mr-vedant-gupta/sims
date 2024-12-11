// Import statements remain the same as in your provided code
package main

//go:generate core generate -add-types

import (
	"fmt"

	"cogentcore.org/core/base/randx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/enums"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
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

func main() {
	sim := &Sim{}
	sim.New()
	sim.ConfigAll()
	sim.RunGUI()
}

// ParamSets is the default set of parameters.
// Base is always applied, and others can be optionally
// selected to apply on top of that.
var ParamSets = params.Sets{
	"Base": {
		{Sel: "Path", Desc: "no extra learning factors",
			Params: params.Params{
				"Path.Learn.Lrate":       "0.01", // slower overall is key
				"Path.Learn.Norm.On":     "false",
				"Path.Learn.Momentum.On": "false",
				"Path.Learn.WtBal.On":    "false",
			}},
		{Sel: "Layer", Desc: "no decay",
			Params: params.Params{
				"Layer.Act.Init.Decay": "0", // key for all layers not otherwise done automatically
			}},
		{Sel: ".BackPath", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates",
			Params: params.Params{
				"Path.WtScale.Rel": "0.2",
			}},
		{Sel: ".BgFixed", Desc: "BG Matrix -> GP wiring",
			Params: params.Params{
				"Path.Learn.Learn": "false",
				"Path.WtInit.Mean": "0.8",
				"Path.WtInit.Var":  "0",
				"Path.WtInit.Sym":  "false",
			}},
		{Sel: ".RWPath", Desc: "Reward prediction -- into PVi",
			Params: params.Params{
				"Path.Learn.Lrate": "0.02",
				"Path.WtInit.Mean": "0",
				"Path.WtInit.Var":  "0",
				"Path.WtInit.Sym":  "false",
			}},
		{Sel: "#Rew", Desc: "Reward layer -- no clamp limits",
			Params: params.Params{
				"Layer.Act.Clamp.Range.Min": "-1",
				"Layer.Act.Clamp.Range.Max": "1",
			}},
		{Sel: ".PFCMntDToOut", Desc: "PFC MntD -> PFC Out fixed",
			Params: params.Params{
				"Path.Learn.Learn": "false",
				"Path.WtInit.Mean": "0.8",
				"Path.WtInit.Var":  "0",
				"Path.WtInit.Sym":  "false",
			}},
		{Sel: ".FmPFCOutD", Desc: "PFC OutD needs to be strong b/c avg act says weak",
			Params: params.Params{
				"Path.WtScale.Abs": "4",
			}},
		{Sel: ".PFCFixed", Desc: "Input -> PFC",
			Params: params.Params{
				"Path.Learn.Learn": "false",
				"Path.WtInit.Mean": "0.8",
				"Path.WtInit.Var":  "0",
				"Path.WtInit.Sym":  "false",
			}},
		{Sel: ".MatrixPath", Desc: "Matrix learning",
			Params: params.Params{
				"Path.Learn.Lrate":         "0.04", // .04 > .1 > .02
				"Path.WtInit.Var":          "0.1",
				"Path.Trace.GateNoGoPosLR": "1",    // 0.1 default
				"Path.Trace.NotGatedLR":    "0.7",  // 0.7 default
				"Path.Trace.Decay":         "1.0",  // 1.0 default
				"Path.Trace.AChDecay":      "0.0",  // not useful even at .1, surprising..
				"Path.Trace.Deriv":         "true", // true default, better than false
			}},
		{Sel: ".MatrixLayer", Desc: "exploring these options",
			Params: params.Params{
				"Layer.Act.XX1.Gain":       "100",
				"Layer.Inhib.Layer.Gi":     "2.2", // 2.2 > 1.8 > 2.4
				"Layer.Inhib.Layer.FB":     "1",   // 1 > .5
				"Layer.Inhib.Pool.On":      "true",
				"Layer.Inhib.Pool.Gi":      "2.1", // def 1.9
				"Layer.Inhib.Pool.FB":      "0",
				"Layer.Inhib.Self.On":      "true",
				"Layer.Inhib.Self.Gi":      "0.4", // def 0.3
				"Layer.Inhib.ActAvg.Init":  "0.05",
				"Layer.Inhib.ActAvg.Fixed": "true",
			}},
		{Sel: "#GPiThal", Desc: "defaults also set automatically by layer but included here just to be sure",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi":     "1.8", // 1.8 > 2.0
				"Layer.Inhib.Layer.FB":     "1",   // 1.0 > 0.5
				"Layer.Inhib.Pool.On":      "false",
				"Layer.Inhib.ActAvg.Init":  ".2",
				"Layer.Inhib.ActAvg.Fixed": "true",
				"Layer.Act.Dt.GTau":        "3",
				"Layer.GPiGate.GeGain":     "3",
				"Layer.GPiGate.NoGo":       "1.25", // was 1 default
				"Layer.GPiGate.Thr":        "0.25", // .2 default
			}},
		{Sel: "#GPeNoGo", Desc: "GPe is a regular layer -- needs special params",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi":     "2.4", // 2.4 > 2.2 > 1.8 > 2.6
				"Layer.Inhib.Layer.FB":     "0.5",
				"Layer.Inhib.Layer.FBTau":  "3", // otherwise a bit jumpy
				"Layer.Inhib.Pool.On":      "false",
				"Layer.Inhib.ActAvg.Init":  ".2",
				"Layer.Inhib.ActAvg.Fixed": "true",
			}},
		{Sel: ".PFC", Desc: "pfc defaults",
			Params: params.Params{
				"Layer.Inhib.Layer.On":     "false",
				"Layer.Inhib.Pool.On":      "true",
				"Layer.Inhib.Pool.Gi":      "1.8",
				"Layer.Inhib.Pool.FB":      "1",
				"Layer.Inhib.ActAvg.Init":  "0.2",
				"Layer.Inhib.ActAvg.Fixed": "true",
			}},
		{Sel: "#Input", Desc: "Basic params",
			Params: params.Params{
				"Layer.Inhib.ActAvg.Init":  "0.25",
				"Layer.Inhib.ActAvg.Fixed": "true",
			}},
		{Sel: "#Output", Desc: "Basic params",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi":     "2",
				"Layer.Inhib.Layer.FB":     "0.5",
				"Layer.Inhib.ActAvg.Init":  "0.25",
				"Layer.Inhib.ActAvg.Fixed": "true",
			}},
		{Sel: "#InputToOutput", Desc: "weaker",
			Params: params.Params{
				"Path.WtScale.Rel": "0.5",
			}},
		{Sel: "#Hidden", Desc: "Basic params",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "2",
				"Layer.Inhib.Layer.FB": "0.5",
			}},
		{Sel: "#SNc", Desc: "allow negative",
			Params: params.Params{
				"Layer.Act.Clamp.Range.Min": "-1",
				"Layer.Act.Clamp.Range.Max": "1",
			}},
		{Sel: "#RWPred", Desc: "keep it guessing",
			Params: params.Params{
				"Layer.RW.PredRange.Min": "0.02", // single most important param!  was .01 -- need penalty..
				"Layer.RW.PredRange.Max": "0.95",
			}},
	},
}

// Config has config parameters related to running the sim
type Config struct {
	// total number of runs to do when running Train
	NRuns int `default:"10" min:"1"`

	// total number of epochs per run
	NEpochs int `default:"200"`

	// total number of trials per epochs per run
	NTrials int `default:"100"`

	// stop run after this number of perfect, zero-error epochs.
	NZero int `default:"5"`

	// how often to run through all the test patterns, in terms of training epochs.
	// can use 0 or -1 for no testing.
	TestInterval int `default:"-1"`
}

// Sim encapsulates the entire simulation model.
type Sim struct {

	// BurstDaGain is the strength of dopamine bursts: 1 default -- reduce for PD OFF, increase for PD ON
	BurstDaGain float32

	// DipDaGain is the strength of dopamine dips: 1 default -- reduce to simulate D2 agonists
	DipDaGain float32

	// Use Entropy measures to modulate the learning rate
	ModLearnRate bool

	// A binary switch for the entropy measure to use (see CalcEntropy)
	EntropyMeasureType bool

	// Config contains misc configuration parameters for running the sim
	Config Config `new-window:"+" display:"no-inline"`

	// the network -- click to view / edit parameters for layers, paths, etc
	Net *leabra.Network `new-window:"+" display:"no-inline"`

	// network parameter management
	Params emer.NetParams `display:"add-fields"`

	// contains looper control loops for running sim
	Loops *looper.Stacks `new-window:"+" display:"no-inline"`

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
	ss.Net = leabra.NewNetwork("SIR")
	ss.Params.Config(ParamSets, "", "", ss.Net)
	ss.Stats.Init()
	ss.Stats.SetInt("Expt", 0)
	ss.RandSeeds.Init(100) // max 100 runs
	ss.InitRandSeed(0)
	ss.Context.Defaults()
}

func (ss *Sim) Defaults() {
	ss.BurstDaGain = 1
	ss.DipDaGain = 1
	ss.ModLearnRate = false
	ss.EntropyMeasureType = false
}

////////////////////////////////////////////////////////////////////////////////
// 		Configs

// ConfigAll configures all the elements using the standard functions
func (ss *Sim) ConfigAll() {
	ss.ConfigEnv()
	ss.ConfigNet(ss.Net)
	ss.ConfigLogs()
	ss.ConfigLoops()
}

func (ss *Sim) ConfigEnv() {
	// Can be called multiple times -- don't re-create
	var trn, tst *SIREnv
	if len(ss.Envs) == 0 {
		trn = &SIREnv{}
		tst = &SIREnv{}
	} else {
		trn = ss.Envs.ByMode(etime.Train).(*SIREnv)
		tst = ss.Envs.ByMode(etime.Test).(*SIREnv)
	}

	// note: names must be standard here!
	trn.Name = etime.Train.String()
	trn.SetNStim(4)
	trn.RewVal = 1
	trn.NoRewVal = 0
	trn.Trial.Max = ss.Config.NTrials

	tst.Name = etime.Test.String()
	tst.SetNStim(4)
	tst.RewVal = 1
	tst.NoRewVal = 0
	tst.Trial.Max = ss.Config.NTrials

	trn.Init(0)
	tst.Init(0)

	// note: names must be in place when adding
	ss.Envs.Add(trn, tst)
}

func (ss *Sim) ConfigNet(net *leabra.Network) {
	net.SetRandSeed(ss.RandSeeds[0]) // init new separate random seed, using run = 0

	rew, rp, da := net.AddRWLayers("", 2)
	da.Name = "SNc"

	inp := net.AddLayer2D("Input", 1, 4, leabra.InputLayer)
	ctrl := net.AddLayer2D("CtrlInput", 1, 5, leabra.InputLayer)
	out := net.AddLayer2D("Output", 1, 4, leabra.TargetLayer)
	hid := net.AddLayer2D("Hidden", 7, 7, leabra.SuperLayer)

	// args: nY, nMaint, nOut, nNeurBgY, nNeurBgX, nNeurPfcY, nNeurPfcX
	mtxGo, mtxNoGo, gpe, gpi, cin, pfcMnt, pfcMntD, pfcOut, pfcOutD := net.AddPBWM("", 4, 2, 2, 1, 5, 1, 4)
	_ = gpe
	_ = gpi
	_ = pfcMnt
	_ = pfcMntD
	_ = pfcOut
	_ = cin

	cin.CIN.RewLays.Add(rew.Name, rp.Name)

	full := paths.NewFull()
	fmin := paths.NewRect()
	fmin.Size.Set(1, 1)
	fmin.Scale.Set(1, 1)
	fmin.Wrap = true

	net.ConnectLayers(ctrl, rp, full, leabra.RWPath)
	net.ConnectLayers(pfcMntD, rp, full, leabra.RWPath)
	net.ConnectLayers(pfcOutD, rp, full, leabra.RWPath)

	net.ConnectLayers(ctrl, mtxGo, fmin, leabra.MatrixPath)
	net.ConnectLayers(ctrl, mtxNoGo, fmin, leabra.MatrixPath)
	pt := net.ConnectLayers(inp, pfcMnt, fmin, leabra.ForwardPath)
	pt.AddClass("PFCFixed")

	net.ConnectLayers(inp, hid, full, leabra.ForwardPath)
	net.ConnectLayers(ctrl, hid, full, leabra.ForwardPath)
	net.BidirConnectLayers(hid, out, full)
	pt = net.ConnectLayers(pfcOutD, hid, full, leabra.ForwardPath)
	pt.AddClass("FmPFCOutD")
	pt = net.ConnectLayers(pfcOutD, out, full, leabra.ForwardPath)
	pt.AddClass("FmPFCOutD")
	net.ConnectLayers(inp, out, full, leabra.ForwardPath)

	inp.PlaceAbove(rew)
	out.PlaceRightOf(inp, 2)
	ctrl.PlaceBehind(inp, 2)
	hid.PlaceBehind(ctrl, 2)
	mtxGo.PlaceRightOf(rew, 2)
	pfcMnt.PlaceRightOf(out, 2)

	net.Build()
	net.Defaults()

	da.AddAllSendToBut() // send dopamine to all layers..
	gpi.SendPBWMParams()

	ss.ApplyParams()
	net.InitWeights()
}

func (ss *Sim) ApplyParams() {
	if ss.Loops != nil {
		trn := ss.Loops.Stacks[etime.Train]
		trn.Loops[etime.Run].Counter.Max = ss.Config.NRuns
		trn.Loops[etime.Epoch].Counter.Max = ss.Config.NEpochs
	}
	ss.Params.SetAll()

	matg := ss.Net.LayerByName("MatrixGo")
	matn := ss.Net.LayerByName("MatrixNoGo")

	matg.Matrix.BurstGain = ss.BurstDaGain
	matg.Matrix.DipGain = ss.DipDaGain
	matn.Matrix.BurstGain = ss.BurstDaGain
	matn.Matrix.DipGain = ss.DipDaGain
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
	ls := looper.NewStacks()

	trls := ss.Config.NTrials

	ls.AddStack(etime.Train).
		AddTime(etime.Run, ss.Config.NRuns).
		AddTime(etime.Epoch, ss.Config.NEpochs).
		AddTime(etime.Trial, trls).
		AddTime(etime.Cycle, 100)

	ls.AddStack(etime.Test).
		AddTime(etime.Epoch, 1).
		AddTime(etime.Trial, trls).
		AddTime(etime.Cycle, 100)

	leabra.LooperStdPhases(ls, &ss.Context, ss.Net, 75, 99)                // plus phase timing
	leabra.LooperSimCycleAndLearn(ls, ss.Net, &ss.Context, &ss.ViewUpdate) // std algo code

	ls.Stacks[etime.Train].OnInit.Add("Init", func() { ss.Init() })

	for m, _ := range ls.Stacks {
		stack := ls.Stacks[m]
		stack.Loops[etime.Trial].OnStart.Add("ApplyInputs", func() {
			ss.ApplyInputs()
		})
	}

	ls.Loop(etime.Train, etime.Run).OnStart.Add("NewRun", ss.NewRun)

	ls.Loop(etime.Train, etime.Run).OnEnd.Add("RunDone", func() {
		if ss.Stats.Int("Run") >= ss.Config.NRuns-1 {
			expt := ss.Stats.Int("Expt")
			ss.Stats.SetInt("Expt", expt+1)
		}
	})

	stack := ls.Stacks[etime.Train]
	cyc, _ := stack.Loops[etime.Cycle]
	plus := cyc.EventByName("MinusPhase:End")
	plus.OnEvent.InsertBefore("MinusPhase:End", "ApplyReward", func() bool {
		ss.ApplyReward(true)
		return true
	})

	// Train stop early condition
	ls.Loop(etime.Train, etime.Epoch).IsDone.AddBool("NZeroStop", func() bool {
		// This is calculated in TrialStats
		stopNz := ss.Config.NZero
		if stopNz <= 0 {
			stopNz = 2
		}
		curNZero := ss.Stats.Int("NZero")
		stop := curNZero >= stopNz
		return stop
	})

	// Add Testing
	trainEpoch := ls.Loop(etime.Train, etime.Epoch)
	trainEpoch.OnStart.Add("TestAtInterval", func() {
		if (ss.Config.TestInterval > 0) && ((trainEpoch.Counter.Cur+1)%ss.Config.TestInterval == 0) {
			// Note the +1 so that it doesn't occur at the 0th timestep.
			ss.TestAll()
		}
	})

	/////////////////////////////////////////////
	// Logging

	ls.Loop(etime.Test, etime.Epoch).OnEnd.Add("LogTestErrors", func() {
		leabra.LogTestErrors(&ss.Logs)
	})
	ls.AddOnEndToAll("Log", func(mode, time enums.Enum) {
		ss.Log(mode.(etime.Modes), time.(etime.Times))
	})
	leabra.LooperResetLogBelow(ls, &ss.Logs)
	ls.Loop(etime.Train, etime.Run).OnEnd.Add("RunStats", func() {
		ss.Logs.RunStats("PctCor", "FirstZero", "LastZero")
	})

	////////////////////////////////////////////
	// GUI

	leabra.LooperUpdateNetView(ls, &ss.ViewUpdate, ss.Net, ss.NetViewCounters)
	leabra.LooperUpdatePlots(ls, &ss.GUI)
	ls.Stacks[etime.Train].OnInit.Add("GUI-Init", func() { ss.GUI.UpdateWindow() })
	ls.Stacks[etime.Test].OnInit.Add("GUI-Init", func() { ss.GUI.UpdateWindow() })

	ss.Loops = ls
}

// ApplyInputs applies input patterns from given environment.
func (ss *Sim) ApplyInputs() {
	ctx := &ss.Context
	net := ss.Net
	ev := ss.Envs.ByMode(ctx.Mode).(*SIREnv)
	ev.Step()

	lays := net.LayersByType(leabra.InputLayer, leabra.TargetLayer)
	net.InitExt()
	ss.Stats.SetString("TrialName", ev.String())
	for _, lnm := range lays {
		if lnm == "Rew" {
			continue
		}
		ly := ss.Net.LayerByName(lnm)
		pats := ev.State(ly.Name)
		if pats != nil {
			ly.ApplyExt(pats)
		}
	}
}

// CalcEntropy computes the entropy based on activations.
func (ss *Sim) CalcEntropy() float32 {
	if !ss.ModLearnRate { // do not modify learning rate
		return 1
	}
	TmpVals := []float32{}
	ent := float32(0)
	if ss.EntropyMeasureType { // first entropy measure: sum of activations in hidden layer
		hid := ss.Net.LayerByName("Hidden")
		hid.UnitValues(&TmpVals, "AvgM", -1)
		maxAct := float32(0)
		for i := range TmpVals {
			maxAct = math32.Max(maxAct, TmpVals[i])
			ent += TmpVals[i]
		}
		// Clamp ent to a desirable range. If the max act among hidden units is too low, set ent to max.
		ent = ent / 10
		if maxAct < 0.02 || ent > 4 {
			ent = 4
		} else if ent < 0.25 {
			ent = 0.25
		}
	} else { // GPiThal entropy calculation for up to 16 units active
		hid := ss.Net.LayerByName("GPiThal")
		hid.UnitValues(&TmpVals, "AvgM", -1)

		// Compute probabilities of activation
		var totalActivation float32
		for _, val := range TmpVals {
			totalActivation += val
		}

		if totalActivation > 0 {
			for _, val := range TmpVals {
				prob := val / totalActivation
				if prob > 0 {
					ent -= prob * math32.Log(prob)
				}
			}
		}

		// Scale entropy to a desirable range
		ent = ent / 2.77 // Normalize entropy for 16 units (max entropy ~2.77 for uniform dist)
		if ent > 4 {
			ent = 4
		} else if ent < 0.25 {
			ent = 0.25
		}
	}
	fmt.Printf("Ent: %g\n", ent)
	return ent
}





// ApplyReward computes reward based on network output and applies it.
// Call at start of 3rd quarter (plus phase).
func (ss *Sim) ApplyReward(train bool) {
	var en *SIREnv
	if train {
		en = ss.Envs.ByMode(etime.Train).(*SIREnv)
	} else {
		en = ss.Envs.ByMode(etime.Test).(*SIREnv)
	}
	if en.Act != Recall1 && en.Act != Recall2 { // only reward on recall trials!
		return
	}
	out := ss.Net.LayerByName("Output")
	mxi := out.Pools[0].Inhib.Act.MaxIndex
	en.SetReward(int(mxi))
	pats := en.State("Rew")
	ly := ss.Net.LayerByName("Rew")
	ly.ApplyExt1DTsr(pats)

	// Control the learning rate in the Matrix and RewPred as a function of "entropy"
	ent := ss.CalcEntropy()
	matg := ss.Net.LayerByName("MatrixGo")
	matng := ss.Net.LayerByName("MatrixNoGo")
	rwpred := ss.Net.LayerByName("RWPred")
	matg.LrateMult(ent)
	matng.LrateMult(ent)
	rwpred.LrateMult(ent)

	if len(matg.RecvPaths) > 0 {
		ss.Stats.SetFloat32("MatrixGoLRate", matg.RecvPaths[0].Learn.Lrate)
	}
	if len(matng.RecvPaths) > 0 {
		ss.Stats.SetFloat32("MatrixNoGoLRate", matng.RecvPaths[0].Learn.Lrate)
	}
	if len(rwpred.RecvPaths) > 0 {
		ss.Stats.SetFloat32("RWPredLRate", rwpred.RecvPaths[0].Learn.Lrate)
	}
}

// NewRun initializes a new run of the model.
func (ss *Sim) NewRun() {
	ctx := &ss.Context
	ss.InitRandSeed(ss.Loops.Loop(etime.Train, etime.Run).Counter.Cur)
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

// TestAll runs through the full set of testing items
func (ss *Sim) TestAll() {
	ss.Envs.ByMode(etime.Test).Init(0)
	ss.Loops.ResetAndRun(etime.Test)
	ss.Loops.Mode = etime.Train // Important to reset Mode back to Train because this is called from within the Train Run.
}

//////////////////////////////////////////////////////////////////////
// 		Stats

// InitStats initializes all the statistics.
func (ss *Sim) InitStats() {
	ss.Stats.SetFloat("SSE", 0.0)
	ss.Stats.SetFloat("DA", 0.0)
	ss.Stats.SetFloat("AbsDA", 0.0)
	ss.Stats.SetFloat("RewPred", 0.0)
	ss.Stats.SetFloat("MatrixGoLRate", 0.0)
	ss.Stats.SetFloat("MatrixNoGoLRate", 0.0)
	ss.Stats.SetFloat("RWPredLRate", 0.0)
	ss.Stats.SetString("TrialName", "")
	ss.Logs.InitErrStats() // inits TrlErr, FirstZero, LastZero, NZero
}

// StatCounters saves current counters to Stats.
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
	ss.ViewUpdate.Text = ss.Stats.Print([]string{"Run", "Epoch", "Trial", "TrialName", "Cycle", "SSE", "TrlErr"})
}

// TrialStats computes the trial-level statistics.
func (ss *Sim) TrialStats() {
	params := fmt.Sprintf("burst: %g, dip: %g", ss.BurstDaGain, ss.DipDaGain)
	ss.Stats.SetString("RunName", params)

	out := ss.Net.LayerByName("Output")

	sse, avgsse := out.MSE(0.5) // 0.5 = per-unit tolerance -- right side of .5
	ss.Stats.SetFloat("SSE", sse)
	ss.Stats.SetFloat("AvgSSE", avgsse)
	if sse > 0 {
		ss.Stats.SetFloat("TrlErr", 1)
	} else {
		ss.Stats.SetFloat("TrlErr", 0)
	}

	snc := ss.Net.LayerByName("SNc")
	ss.Stats.SetFloat32("DA", snc.Neurons[0].Act)
	ss.Stats.SetFloat32("AbsDA", math32.Abs(snc.Neurons[0].Act))
	rp := ss.Net.LayerByName("RWPred")
	ss.Stats.SetFloat32("RewPred", rp.Neurons[0].Act)
}

//////////////////////////////////////////////////////////////////////
// 		Logging

func (ss *Sim) ConfigLogs() {
	ss.Stats.SetString("RunName", ss.Params.RunName(0)) // used for naming logs, stats, etc

	ss.Logs.AddCounterItems(etime.Run, etime.Epoch, etime.Trial, etime.Cycle)
	ss.Logs.AddStatIntNoAggItem(etime.AllModes, etime.AllTimes, "Expt")
	ss.Logs.AddStatStringItem(etime.AllModes, etime.AllTimes, "RunName")
	ss.Logs.AddStatStringItem(etime.AllModes, etime.Trial, "TrialName")

	ss.Logs.AddPerTrlMSec("PerTrlMSec", etime.Run, etime.Epoch, etime.Trial)

	ss.Logs.AddStatAggItem("SSE", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("AvgSSE", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddErrStatAggItems("TrlErr", etime.Run, etime.Epoch, etime.Trial)

	ss.Logs.AddStatAggItem("DA", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("AbsDA", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("RewPred", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("MatrixGoLRate", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("MatrixNoGoLRate", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("RWPredLRate", etime.Run, etime.Epoch, etime.Trial)

	ss.Logs.PlotItems("PctErr", "AbsDA", "RewPred")
	ss.Logs.CreateTables()
	ss.Logs.SetContext(&ss.Stats, ss.Net)
	// don't plot certain combinations we don't use
	ss.Logs.NoPlot(etime.Train, etime.Cycle)
	ss.Logs.NoPlot(etime.Test, etime.Cycle)
	ss.Logs.NoPlot(etime.Test, etime.Trial)
	ss.Logs.NoPlot(etime.Test, etime.Run)
	ss.Logs.SetMeta(etime.Train, etime.Run, "LegendCol", "RunName")
}

// Log is the main logging function.
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
	}

	ss.Logs.LogRow(mode, time, row) // also logs to file, etc

	if mode == etime.Test {
		ss.GUI.UpdateTableView(etime.Test, etime.Trial)
	}
}

//////////////////////////////////////////////////////////////////////
// 		GUI

// ConfigGUI configures the GUI interface for this simulation.
func (ss *Sim) ConfigGUI() {
	title := "SIR"
	ss.GUI.MakeBody(ss, "sir", title, `sir illustrates the dynamic gating of information into PFC active maintenance, by the basal ganglia (BG). It uses a simple Store-Ignore-Recall (SIR) task, where the BG system learns via phasic dopamine signals and trial-and-error exploration, discovering what needs to be stored, ignored, and recalled as a function of reinforcement of correct behavior, and learned reinforcement of useful working memory representations. See <a href="https://github.com/CompCogNeuro/sims/blob/master/ch9/sir/README.md">README.md on GitHub</a>.</p>`)
	ss.GUI.CycleUpdateInterval = 10

	nv := ss.GUI.AddNetView("Network")
	nv.Options.MaxRecs = 300
	nv.Options.Raster.Max = 100
	nv.SetNet(ss.Net)
	nv.Options.PathWidth = 0.003
	ss.ViewUpdate.Config(nv, etime.GammaCycle, etime.GammaCycle)
	ss.GUI.ViewUpdate = &ss.ViewUpdate
	nv.Current()

	// nv.SceneXYZ().Camera.Pose.Pos.Set(0, 1.15, 2.25)
	// nv.SceneXYZ().Camera.LookAt(math32.Vector3{0, -0.15, 0}, math32.Vector3{0, 1, 0})

	ss.GUI.AddPlots(title, &ss.Logs)

	ss.GUI.AddTableView(&ss.Logs, etime.Test, etime.Trial)

	ss.GUI.FinalizeGUI(false)
}

func (ss *Sim) MakeToolbar(p *tree.Plan) {
	ss.GUI.AddLooperCtrl(p, ss.Loops)

	tree.Add(p, func(w *core.Separator) {})
	ss.GUI.AddToolbarItem(p, egui.ToolbarItem{Label: "Reset RunLog",
		Icon:    icons.Reset,
		Tooltip: "Reset the accumulated log of all Runs, which are tagged with the ParamSet used",
		Active:  egui.ActiveAlways,
		Func: func() {
			ss.Logs.ResetLog(etime.Train, etime.Run)
			ss.GUI.UpdatePlot(etime.Train, etime.Run)
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
			core.TheApp.OpenURL("https://github.com/CompCogNeuro/sims/blob/main/ch9/sir/README.md")
		},
	})
}

func (ss *Sim) RunGUI() {
	ss.Init()
	ss.ConfigGUI()
	ss.GUI.Body.RunMainWindow()
}