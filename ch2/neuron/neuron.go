// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
neuron: This simulation illustrates the basic properties of neural spiking and
rate-code activation, reflecting a balance of excitatory and inhibitory
influences (including leak and synaptic inhibition).
*/
package main

import (
	"fmt"
	"log"
	"strconv"

	"cogentcore.org/core/events"
	"cogentcore.org/core/gi"
	"cogentcore.org/core/giv"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/ki"
	"cogentcore.org/core/styles"
	"github.com/emer/emergent/v2/emer"
	"github.com/emer/emergent/v2/netview"
	"github.com/emer/emergent/v2/params"
	"github.com/emer/etable/v2/eplot"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
	"github.com/emer/leabra/v2/leabra"
	"github.com/emer/leabra/v2/spike"
)

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// ParamSets is the default set of parameters -- Base is always applied, and others can be optionally
// selected to apply on top of that
var ParamSets = params.Sets{
	"Base": {Desc: "these are the best params", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Prjn", Desc: "no learning",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
				}},
			{Sel: "Layer", Desc: "generic params for all layers: lower gain, slower, soft clamp",
				Params: params.Params{
					"Layer.Inhib.Layer.On":  "false",
					"Layer.Act.XX1.Gain":    "30",
					"Layer.Act.XX1.NVar":    "0.01",
					"Layer.Act.Init.Vm":     "0.3",
					"Layer.Act.Noise.Dist":  "Gaussian",
					"Layer.Act.Noise.Var":   "0",
					"Layer.Act.Noise.Type":  "GeNoise",
					"Layer.Act.Noise.Fixed": "false",
				}},
		},
	}},
}

// Sim encapsulates the entire simulation model, and we define all the
// functionality as methods on this struct.  This structure keeps all relevant
// state information organized and available without having to pass everything around
// as arguments to methods, and provides the core GUI interface (note the view tags
// for the fields which provide hints to how things should be displayed).
type Sim struct {
	// use discrete spiking equations -- otherwise use Noisy X-over-X-plus-1 rate code activation function
	Spike bool
	// excitatory conductance multiplier -- determines overall value of Ge which drives neuron to be more excited -- pushes up over threshold to fire if strong enough
	GbarE float32 `min:"0" step:"0.01" def:"0.3"`
	// leak conductance -- determines overall value of Gl which drives neuron to be less excited (inhibited) -- pushes back to resting membrane potential
	GbarL float32 `min:"0" step:"0.01" def:"0.3"`
	// excitatory reversal (driving) potential -- determines where excitation pushes Vm up to
	ErevE float32 `min:"0" max:"1" step:"0.01" def:"1"`
	// leak reversal (driving) potential -- determines where excitation pulls Vm down to
	ErevL float32 `min:"0" max:"1" step:"0.01" def:"0.3"`
	// the variance parameter for Gaussian noise added to unit activations on every cycle
	Noise float32 `min:"0" step:"0.01"`
	// apply sodium-gated potassium adaptation mechanisms that cause the neuron to reduce spiking over time
	KNaAdapt bool
	// total number of cycles to run
	NCycles int `min:"10" def:"200"`
	// when does excitatory input into neuron come on?
	OnCycle int `min:"0" def:"10"`
	// when does excitatory input into neuron go off?
	OffCycle int `min:"0" def:"160"`
	// how often to update display (in cycles)
	UpdateInterval int `min:"1" def:"10"`
	// the network -- click to view / edit parameters for layers, prjns, etc
	Net *leabra.Network `view:"no-inline"`
	// parameters for spiking funcion
	SpikeParams spike.ActParams `view:"no-inline"`
	// testing trial-level log data -- click to see record of network's response to each input
	TstCycLog *etable.Table `view:"no-inline"`
	// plot of measured spike rate vs. noisy X/X+1 rate function
	SpikeVsRateLog *etable.Table `view:"no-inline"`
	// full collection of param sets -- not really interesting for this model
	Params params.Sets `view:"no-inline"`

	// current cycle of updating
	Cycle int `inactive:"+"`

	// main GUI body
	Body *gi.Body `view:"-"`
	// the network viewer
	NetView *netview.NetView `view:"-"`
	// the test-trial plot
	TstCycPlot *eplot.Plot2D `view:"-"`
	// the spike vs. rate plot
	SpikeVsRatePlot *eplot.Plot2D `view:"-"`
	// true if sim is running
	IsRunning bool `view:"-"`
	// flag to stop running
	StopNow bool `view:"-"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// New creates new blank elements and initializes defaults
func (ss *Sim) New() {
	ss.Net = &leabra.Network{}
	ss.TstCycLog = &etable.Table{}
	ss.SpikeVsRateLog = &etable.Table{}
	ss.Params = ParamSets
	ss.Defaults()
	ss.SpikeParams.Defaults()
}

// Defaults sets default params
func (ss *Sim) Defaults() {
	ss.UpdateInterval = 10
	ss.Cycle = 0
	ss.Spike = true
	ss.GbarE = 0.3
	ss.GbarL = 0.3
	ss.ErevE = 1
	ss.ErevL = 0.3
	ss.Noise = 0
	ss.KNaAdapt = true
	ss.NCycles = 200
	ss.OnCycle = 10
	ss.OffCycle = 160
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Configs

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.ConfigNet(ss.Net)
	ss.ConfigTstCycLog(ss.TstCycLog)
	ss.ConfigSpikeVsRateLog(ss.SpikeVsRateLog)
}

func (ss *Sim) ConfigNet(net *leabra.Network) {
	net.InitName(net, "Neuron")
	net.AddLayer2D("Neuron", 1, 1, emer.Hidden)

	net.Defaults()
	ss.SetParams("Network", false) // only set Network params
	err := net.Build()
	if err != nil {
		log.Println(err)
		return
	}
	ss.InitWts(net)
}

// InitWts loads the saved weights
func (ss *Sim) InitWts(net *leabra.Network) {
	net.InitWts()
}

////////////////////////////////////////////////////////////////////////////////
// 	    Init, utils

// Init restarts the run, and initializes everything, including network weights
// and resets the epoch log table
func (ss *Sim) Init() {
	ss.Cycle = 0
	ss.InitWts(ss.Net)
	ss.StopNow = false
	ss.SetParams("", false) // all sheets
	ss.UpdateView(-1)
	if ss.NetView != nil && ss.NetView.IsVisible() {
		ss.NetView.RecordSyns()
	}
}

// Counters returns a string of the current counter state
// use tabs to achieve a reasonable formatting overall
// and add a few tabs at the end to allow for expansion..
func (ss *Sim) Counters() string {
	return fmt.Sprintf("Cycle:\t%d\t\t\t", ss.Cycle)
}

func (ss *Sim) UpdateView(cyc int) {
	if ss.NetView != nil && ss.NetView.IsVisible() {
		ss.NetView.Record(ss.Counters(), cyc)
		// note: essential to use Go version of update when called from another goroutine
		ss.NetView.Update() // note: using counters is significantly slower..
	}
}

////////////////////////////////////////////////////////////////////////////////
// 	    Running the Network, starting bottom-up..

// RunCycles updates neuron over specified number of cycles
func (ss *Sim) RunCycles() {
	ss.Init()
	ss.StopNow = false
	ss.Net.InitActs()
	ss.SetParams("", false)
	ly := ss.Net.LayerByName("Neuron").(leabra.LeabraLayer).AsLeabra()
	nrn := &(ly.Neurons[0])
	inputOn := false
	for cyc := 0; cyc < ss.NCycles; cyc++ {
		ss.Cycle = cyc
		switch cyc {
		case ss.OnCycle:
			inputOn = true
		case ss.OffCycle:
			inputOn = false
		}
		nrn.Noise = float32(ly.Act.Noise.Gen(-1))
		if inputOn {
			nrn.Ge = 1
		} else {
			nrn.Ge = 0
		}
		nrn.Ge += nrn.Noise // GeNoise
		nrn.Gi = 0
		if ss.Spike {
			ss.SpikeUpdate(ss.Net, inputOn)
		} else {
			ss.RateUpdate(ss.Net, inputOn)
		}
		ss.LogTstCyc(ss.TstCycLog, ss.Cycle)
		if ss.Cycle%ss.UpdateInterval == 0 {
			ss.UpdateView(ss.Cycle)
		}
		if ss.StopNow {
			break
		}
	}
	ss.UpdateView(ss.Cycle)
}

// RateUpdate updates the neuron in rate-code mode
// this just calls the relevant activation code directly, bypassing most other stuff.
func (ss *Sim) RateUpdate(nt *leabra.Network, inputOn bool) {
	ly := ss.Net.LayerByName("Neuron").(leabra.LeabraLayer).AsLeabra()
	nrn := &(ly.Neurons[0])
	ly.Act.VmFmG(nrn)
	ly.Act.ActFmG(nrn)
	nrn.Ge = nrn.Ge * ly.Act.Gbar.E // display effective Ge
}

// SpikeUpdate updates the neuron in spiking mode
// which is just computed directly as spiking is not yet implemented in main codebase
func (ss *Sim) SpikeUpdate(nt *leabra.Network, inputOn bool) {
	ly := ss.Net.LayerByName("Neuron").(leabra.LeabraLayer).AsLeabra()
	nrn := &(ly.Neurons[0])
	ss.SpikeParams.SpikeVmFmG(nrn)
	ss.SpikeParams.SpikeActFmVm(nrn)
	nrn.Ge = nrn.Ge * ly.Act.Gbar.E // display effective Ge
}

// Stop tells the sim to stop running
func (ss *Sim) Stop() {
	ss.StopNow = true
}

// SpikeVsRate runs comparison between spiking vs. rate-code
func (ss *Sim) SpikeVsRate() {
	row := 0
	nsamp := 100
	// ss.KNaAdapt = false
	for gbarE := 0.1; gbarE <= 0.7; gbarE += 0.025 {
		ss.GbarE = float32(gbarE)
		spike := float64(0)
		ss.Noise = 0.1 // RunCycles calls SetParams to set this
		ss.Spike = true
		for ns := 0; ns < nsamp; ns++ {
			ss.RunCycles()
			if ss.StopNow {
				break
			}
			act := ss.TstCycLog.CellFloat("Act", 159)
			spike += act
		}
		rate := float64(0)
		ss.Spike = false
		// ss.Noise = 0 // doesn't make much diff
		for ns := 0; ns < nsamp; ns++ {
			ss.RunCycles()
			if ss.StopNow {
				break
			}
			act := ss.TstCycLog.CellFloat("Act", 159)
			rate += act
		}
		if ss.StopNow {
			break
		}
		spike /= float64(nsamp)
		rate /= float64(nsamp)
		ss.LogSpikeVsRate(ss.SpikeVsRateLog, row, gbarE, spike, rate)
		row++
	}
	ss.Defaults()
	ss.SpikeVsRatePlot.Update()
}

/////////////////////////////////////////////////////////////////////////
//   Params setting

// SetParams sets the params for "Base" and then current ParamSet.
// If sheet is empty, then it applies all avail sheets (e.g., Network, Sim)
// otherwise just the named sheet
// if setMsg = true then we output a message for each param that was set.
func (ss *Sim) SetParams(sheet string, setMsg bool) error {
	if sheet == "" {
		// this is important for catching typos and ensuring that all sheets can be used
		ss.Params.ValidateSheets([]string{"Network", "Sim"})
	}
	err := ss.SetParamsSet("Base", sheet, setMsg)
	ly := ss.Net.LayerByName("Neuron").(leabra.LeabraLayer).AsLeabra()
	ly.Act.Gbar.E = float32(ss.GbarE)
	ly.Act.Gbar.L = float32(ss.GbarL)
	ly.Act.Erev.E = float32(ss.ErevE)
	ly.Act.Erev.L = float32(ss.ErevL)
	ly.Act.Noise.Var = float64(ss.Noise)
	ly.Act.KNa.On = ss.KNaAdapt
	ly.Act.Update()
	ss.SpikeParams.ActParams = ly.Act // keep sync'd
	ss.SpikeParams.KNa.On = ss.KNaAdapt
	return err
}

// SetParamsSet sets the params for given params.Set name.
// If sheet is empty, then it applies all avail sheets (e.g., Network, Sim)
// otherwise just the named sheet
// if setMsg = true then we output a message for each param that was set.
func (ss *Sim) SetParamsSet(setNm string, sheet string, setMsg bool) error {
	pset, err := ss.Params.SetByNameTry(setNm)
	if err != nil {
		return err
	}
	if sheet == "" || sheet == "Network" {
		netp, ok := pset.Sheets["Network"]
		if ok {
			ss.Net.ApplyParams(netp, setMsg)
		}
	}

	if sheet == "" || sheet == "Sim" {
		simp, ok := pset.Sheets["Sim"]
		if ok {
			simp.Apply(ss, setMsg)
		}
	}
	// note: if you have more complex environments with parameters, definitely add
	// sheets for them, e.g., "TrainEnv", "TestEnv" etc
	return err
}

//////////////////////////////////////////////
//  TstCycLog

// LogTstCyc adds data from current cycle to the TstCycLog table.
func (ss *Sim) LogTstCyc(dt *etable.Table, cyc int) {
	if dt.Rows <= cyc {
		dt.SetNumRows(cyc + 1)
	}
	row := cyc

	ly := ss.Net.LayerByName("Neuron").(leabra.LeabraLayer).AsLeabra()
	nrn := &(ly.Neurons[0])

	dt.SetCellFloat("Cycle", row, float64(cyc))
	dt.SetCellFloat("Ge", row, float64(nrn.Ge))
	dt.SetCellFloat("Inet", row, float64(nrn.Inet))
	dt.SetCellFloat("Vm", row, float64(nrn.Vm))
	dt.SetCellFloat("Act", row, float64(nrn.Act))
	dt.SetCellFloat("Spike", row, float64(nrn.Spike))
	dt.SetCellFloat("Gk", row, float64(nrn.Gk))
	dt.SetCellFloat("ISI", row, float64(nrn.ISI))
	dt.SetCellFloat("AvgISI", row, float64(nrn.ISIAvg))

	// note: essential to use Go version of update when called from another goroutine
	if cyc%ss.UpdateInterval == 0 {
		ss.TstCycPlot.Update()
	}
}

func (ss *Sim) ConfigTstCycLog(dt *etable.Table) {
	dt.SetMetaData("name", "TstCycLog")
	dt.SetMetaData("desc", "Record of testing per cycle")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	nt := ss.NCycles // max cycles
	sch := etable.Schema{
		{"Cycle", etensor.INT64, nil, nil},
		{"Ge", etensor.FLOAT64, nil, nil},
		{"Inet", etensor.FLOAT64, nil, nil},
		{"Vm", etensor.FLOAT64, nil, nil},
		{"Act", etensor.FLOAT64, nil, nil},
		{"Spike", etensor.FLOAT64, nil, nil},
		{"Gk", etensor.FLOAT64, nil, nil},
		{"ISI", etensor.FLOAT64, nil, nil},
		{"AvgISI", etensor.FLOAT64, nil, nil},
	}
	dt.SetFromSchema(sch, nt)
}

func (ss *Sim) ConfigTstCycPlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Neuron Test Cycle Plot"
	plt.Params.XAxisCol = "Cycle"
	plt.SetTable(dt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("Cycle", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Ge", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("Inet", eplot.On, eplot.FixMin, -.2, eplot.FixMax, 1)
	plt.SetColParams("Vm", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("Act", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("Spike", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("Gk", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("ISI", eplot.Off, eplot.FixMin, -2, eplot.FloatMax, 1)
	plt.SetColParams("AvgISI", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
	return plt
}

func (ss *Sim) ResetTstCycPlot() {
	ss.TstCycLog.SetNumRows(0)
	ss.TstCycPlot.Update()
}

//////////////////////////////////////////////
//  SpikeVsRateLog

// LogSpikeVsRate adds data from current cycle to the SpikeVsRateLog table.
func (ss *Sim) LogSpikeVsRate(dt *etable.Table, row int, gbarE, spike, rate float64) {
	if dt.Rows <= row {
		dt.SetNumRows(row + 1)
	}
	dt.SetCellFloat("GBarE", row, gbarE)
	dt.SetCellFloat("Spike", row, spike)
	dt.SetCellFloat("Rate", row, rate)
}

func (ss *Sim) ConfigSpikeVsRateLog(dt *etable.Table) {
	dt.SetMetaData("name", "SpikeVsRateLog")
	dt.SetMetaData("desc", "Record spiking vs. rate-code activation")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	nt := 24 // typical number
	sch := etable.Schema{
		{"GBarE", etensor.FLOAT64, nil, nil},
		{"Spike", etensor.FLOAT64, nil, nil},
		{"Rate", etensor.FLOAT64, nil, nil},
	}
	dt.SetFromSchema(sch, nt)
}

func (ss *Sim) ConfigSpikeVsRatePlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Neuron Spike Vs. Rate-Code Plot"
	plt.Params.XAxisCol = "GBarE"
	plt.SetTable(dt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("GBarE", eplot.Off, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("Spike", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("Rate", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	return plt
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Gui

func (ss *Sim) ConfigNetView(nv *netview.NetView) {
	nv.ViewDefaults()
}

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() {
	b := gi.NewBody("Neuron")
	ss.Body = b

	split := gi.NewSplits(b, "split")

	giv.NewStructView(split, "sv").SetStruct(ss)

	tv := gi.NewTabs(split, "tv")

	nv := netview.NewNetView(tv.NewTab("NetView"))
	nv.Var = "Act"
	nv.SetNet(ss.Net)
	ss.NetView = nv
	ss.ConfigNetView(nv) // add labels etc

	plt := eplot.NewPlot2D(tv.NewTab("TstCycPlot"))
	ss.TstCycPlot = ss.ConfigTstCycPlot(plt, ss.TstCycLog)

	plt = eplot.NewPlot2D(tv.NewTab("SpikeVsRatePlot"))
	ss.SpikeVsRatePlot = ss.ConfigSpikeVsRatePlot(plt, ss.SpikeVsRateLog)

	split.SetSplits(.2, .8)

	b.AddAppBar(func(tb *gi.Toolbar) {
		gi.NewButton(tb).SetText("Init").SetIcon(icons.Update).
			SetTooltip("Initialize everything including network weights, and start over.  Also applies current params.").
			StyleFirst(func(s *styles.Style) {
				s.SetEnabled(!ss.IsRunning)
			}).
			OnClick(func(e events.Event) {
				ss.Init()
			})

		gi.NewButton(tb).SetText("Stop").SetIcon(icons.Stop).
			SetTooltip("Interrupts running.  Hitting Train again will pick back up where it left off.").
			StyleFirst(func(s *styles.Style) {
				s.SetEnabled(ss.IsRunning)
			}).
			OnClick(func(e events.Event) {
				ss.Stop()
			})

		gi.NewButton(tb).SetText("Run cycles").SetIcon(icons.Step).
			SetTooltip("Runs neuron updating over NCycles.").
			StyleFirst(func(s *styles.Style) {
				s.SetEnabled(!ss.IsRunning)
			}).
			OnClick(func(e events.Event) {
				ss.IsRunning = true
				ss.RunCycles()
				ss.IsRunning = false
			})

		gi.NewSeparator(tb)

		gi.NewButton(tb).SetText("Reset plot").SetIcon(icons.Reset).
			SetTooltip("Reset TstCycPlot.").
			StyleFirst(func(s *styles.Style) {
				s.SetEnabled(!ss.IsRunning)
			}).
			OnClick(func(e events.Event) {
				ss.ResetTstCycPlot()
			})

		gi.NewButton(tb).SetText("Spike vs Rate").SetIcon(icons.PlayArrow).
			SetTooltip("Runs Spike vs Rate test.").
			StyleFirst(func(s *styles.Style) {
				s.SetEnabled(!ss.IsRunning)
			}).
			OnClick(func(e events.Event) {
				ss.IsRunning = true
				go func() {
					ss.SpikeVsRate()
					ss.IsRunning = false
				}()
			})

		gi.NewButton(tb).SetText("Defaults").SetIcon(icons.Reset).
			SetTooltip("Restore initial default parameters.").
			StyleFirst(func(s *styles.Style) {
				s.SetEnabled(!ss.IsRunning)
			}).
			OnClick(func(e events.Event) {
				ss.Defaults()
				ss.Init()
			})

		gi.NewButton(tb).SetText("README").SetIcon(icons.FileMarkdown).
			SetTooltip("Opens your browser on the README file that contains instructions for how to run this model.").
			OnClick(func(e events.Event) {
				gi.TheApp.OpenURL("https://github.com/CompCogNeuro/sims/blob/master/ch2/neuron/README.md")
			})
	})
}

// These props register Save methods so they can be used
var SimProps = ki.Props{
	"CallMethods": ki.PropSlice{
		{"SaveWeights", ki.Props{
			"desc": "save network weights to file",
			"icon": "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".wts",
				}},
			},
		}},
	},
}

func main() {
	TheSim.New()
	TheSim.Config()

	TheSim.Init()
	TheSim.ConfigGui()
	TheSim.Body.RunMainWindow()
}
