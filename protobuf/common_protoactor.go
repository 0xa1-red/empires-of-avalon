// Package protobuf is generated by protoactor-go/protoc-gen-gograin@0.1.0
package protobuf

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	logmod "github.com/asynkron/protoactor-go/log"
	"google.golang.org/protobuf/proto"
)

var (
	plog = logmod.New(logmod.InfoLevel, "[GRAIN][protobuf]")
	_    = proto.Marshal
	_    = fmt.Errorf
	_    = math.Inf
)

// SetLogLevel sets the log level.
func SetLogLevel(level logmod.Level) {
	plog.SetLevel(level)
}

var xInventoryFactory func() Inventory

// InventoryFactory produces a Inventory
func InventoryFactory(factory func() Inventory) {
	xInventoryFactory = factory
}

// GetInventoryGrainClient instantiates a new InventoryGrainClient with given Identity
func GetInventoryGrainClient(c *cluster.Cluster, id string) *InventoryGrainClient {
	if c == nil {
		panic(fmt.Errorf("nil cluster instance"))
	}
	if id == "" {
		panic(fmt.Errorf("empty id"))
	}
	return &InventoryGrainClient{Identity: id, cluster: c}
}

// GetInventoryKind instantiates a new cluster.Kind for Inventory
func GetInventoryKind(opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(func() actor.Actor {
		return &InventoryActor{
			Timeout: 60 * time.Second,
		}
	}, opts...)
	kind := cluster.NewKind("Inventory", props)
	return kind
}

// GetInventoryKind instantiates a new cluster.Kind for Inventory
func NewInventoryKind(factory func() Inventory, timeout time.Duration, opts ...actor.PropsOption) *cluster.Kind {
	xInventoryFactory = factory
	props := actor.PropsFromProducer(func() actor.Actor {
		return &InventoryActor{
			Timeout: timeout,
		}
	}, opts...)
	kind := cluster.NewKind("Inventory", props)
	return kind
}

// Inventory interfaces the services available to the Inventory
type Inventory interface {
	Init(ctx cluster.GrainContext)
	Terminate(ctx cluster.GrainContext)
	ReceiveDefault(ctx cluster.GrainContext)
	StartBuilding(*StartBuildingRequest, cluster.GrainContext) (*StartBuildingResponse, error)
	Describe(*DescribeInventoryRequest, cluster.GrainContext) (*DescribeInventoryResponse, error)
	Reserve(*ReserveRequest, cluster.GrainContext) (*ReserveResponse, error)
	Persist(*InventoryPersistRequest, cluster.GrainContext) (*InventoryPersistResponse, error)
}

// InventoryGrainClient holds the base data for the InventoryGrain
type InventoryGrainClient struct {
	Identity string
	cluster  *cluster.Cluster
}

// StartBuilding requests the execution on to the cluster with CallOptions
func (g *InventoryGrainClient) StartBuilding(r *StartBuildingRequest, opts ...cluster.GrainCallOption) (*StartBuildingResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 0, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Inventory", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &StartBuildingResponse{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// Describe requests the execution on to the cluster with CallOptions
func (g *InventoryGrainClient) Describe(r *DescribeInventoryRequest, opts ...cluster.GrainCallOption) (*DescribeInventoryResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 1, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Inventory", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &DescribeInventoryResponse{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// Reserve requests the execution on to the cluster with CallOptions
func (g *InventoryGrainClient) Reserve(r *ReserveRequest, opts ...cluster.GrainCallOption) (*ReserveResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 2, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Inventory", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &ReserveResponse{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// Persist requests the execution on to the cluster with CallOptions
func (g *InventoryGrainClient) Persist(r *InventoryPersistRequest, opts ...cluster.GrainCallOption) (*InventoryPersistResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 3, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Inventory", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &InventoryPersistResponse{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// InventoryActor represents the actor structure
type InventoryActor struct {
	ctx     cluster.GrainContext
	inner   Inventory
	Timeout time.Duration
}

// Receive ensures the lifecycle of the actor for the received message
func (a *InventoryActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started: //pass
	case *cluster.ClusterInit:
		a.ctx = cluster.NewGrainContext(ctx, msg.Identity, msg.Cluster)
		a.inner = xInventoryFactory()
		a.inner.Init(a.ctx)

		if a.Timeout > 0 {
			ctx.SetReceiveTimeout(a.Timeout)
		}
	case *actor.ReceiveTimeout:
		ctx.Poison(ctx.Self())
	case *actor.Stopped:
		a.inner.Terminate(a.ctx)
	case actor.AutoReceiveMessage: // pass
	case actor.SystemMessage: // pass

	case *cluster.GrainRequest:
		switch msg.MethodIndex {
		case 0:
			req := &StartBuildingRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("StartBuilding(StartBuildingRequest) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.StartBuilding(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("StartBuilding(StartBuildingRequest) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)
		case 1:
			req := &DescribeInventoryRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("Describe(DescribeInventoryRequest) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.Describe(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("Describe(DescribeInventoryRequest) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)
		case 2:
			req := &ReserveRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("Reserve(ReserveRequest) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.Reserve(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("Reserve(ReserveRequest) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)
		case 3:
			req := &InventoryPersistRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("Persist(InventoryPersistRequest) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.Persist(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("Persist(InventoryPersistRequest) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)

		}
	default:
		a.inner.ReceiveDefault(a.ctx)
	}
}

var xTimerFactory func() Timer

// TimerFactory produces a Timer
func TimerFactory(factory func() Timer) {
	xTimerFactory = factory
}

// GetTimerGrainClient instantiates a new TimerGrainClient with given Identity
func GetTimerGrainClient(c *cluster.Cluster, id string) *TimerGrainClient {
	if c == nil {
		panic(fmt.Errorf("nil cluster instance"))
	}
	if id == "" {
		panic(fmt.Errorf("empty id"))
	}
	return &TimerGrainClient{Identity: id, cluster: c}
}

// GetTimerKind instantiates a new cluster.Kind for Timer
func GetTimerKind(opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(func() actor.Actor {
		return &TimerActor{
			Timeout: 60 * time.Second,
		}
	}, opts...)
	kind := cluster.NewKind("Timer", props)
	return kind
}

// GetTimerKind instantiates a new cluster.Kind for Timer
func NewTimerKind(factory func() Timer, timeout time.Duration, opts ...actor.PropsOption) *cluster.Kind {
	xTimerFactory = factory
	props := actor.PropsFromProducer(func() actor.Actor {
		return &TimerActor{
			Timeout: timeout,
		}
	}, opts...)
	kind := cluster.NewKind("Timer", props)
	return kind
}

// Timer interfaces the services available to the Timer
type Timer interface {
	Init(ctx cluster.GrainContext)
	Terminate(ctx cluster.GrainContext)
	ReceiveDefault(ctx cluster.GrainContext)
	CreateTimer(*TimerRequest, cluster.GrainContext) (*TimerResponse, error)
	Describe(*DescribeTimerRequest, cluster.GrainContext) (*DescribeTimerResponse, error)
}

// TimerGrainClient holds the base data for the TimerGrain
type TimerGrainClient struct {
	Identity string
	cluster  *cluster.Cluster
}

// CreateTimer requests the execution on to the cluster with CallOptions
func (g *TimerGrainClient) CreateTimer(r *TimerRequest, opts ...cluster.GrainCallOption) (*TimerResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 0, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Timer", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &TimerResponse{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// Describe requests the execution on to the cluster with CallOptions
func (g *TimerGrainClient) Describe(r *DescribeTimerRequest, opts ...cluster.GrainCallOption) (*DescribeTimerResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 1, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Timer", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &DescribeTimerResponse{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// TimerActor represents the actor structure
type TimerActor struct {
	ctx     cluster.GrainContext
	inner   Timer
	Timeout time.Duration
}

// Receive ensures the lifecycle of the actor for the received message
func (a *TimerActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started: //pass
	case *cluster.ClusterInit:
		a.ctx = cluster.NewGrainContext(ctx, msg.Identity, msg.Cluster)
		a.inner = xTimerFactory()
		a.inner.Init(a.ctx)

		if a.Timeout > 0 {
			ctx.SetReceiveTimeout(a.Timeout)
		}
	case *actor.ReceiveTimeout:
		ctx.Poison(ctx.Self())
	case *actor.Stopped:
		a.inner.Terminate(a.ctx)
	case actor.AutoReceiveMessage: // pass
	case actor.SystemMessage: // pass

	case *cluster.GrainRequest:
		switch msg.MethodIndex {
		case 0:
			req := &TimerRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("CreateTimer(TimerRequest) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.CreateTimer(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("CreateTimer(TimerRequest) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)
		case 1:
			req := &DescribeTimerRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("Describe(DescribeTimerRequest) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.Describe(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("Describe(DescribeTimerRequest) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)

		}
	default:
		a.inner.ReceiveDefault(a.ctx)
	}
}

var xAdminFactory func() Admin

// AdminFactory produces a Admin
func AdminFactory(factory func() Admin) {
	xAdminFactory = factory
}

// GetAdminGrainClient instantiates a new AdminGrainClient with given Identity
func GetAdminGrainClient(c *cluster.Cluster, id string) *AdminGrainClient {
	if c == nil {
		panic(fmt.Errorf("nil cluster instance"))
	}
	if id == "" {
		panic(fmt.Errorf("empty id"))
	}
	return &AdminGrainClient{Identity: id, cluster: c}
}

// GetAdminKind instantiates a new cluster.Kind for Admin
func GetAdminKind(opts ...actor.PropsOption) *cluster.Kind {
	props := actor.PropsFromProducer(func() actor.Actor {
		return &AdminActor{
			Timeout: 60 * time.Second,
		}
	}, opts...)
	kind := cluster.NewKind("Admin", props)
	return kind
}

// GetAdminKind instantiates a new cluster.Kind for Admin
func NewAdminKind(factory func() Admin, timeout time.Duration, opts ...actor.PropsOption) *cluster.Kind {
	xAdminFactory = factory
	props := actor.PropsFromProducer(func() actor.Actor {
		return &AdminActor{
			Timeout: timeout,
		}
	}, opts...)
	kind := cluster.NewKind("Admin", props)
	return kind
}

// Admin interfaces the services available to the Admin
type Admin interface {
	Init(ctx cluster.GrainContext)
	Terminate(ctx cluster.GrainContext)
	ReceiveDefault(ctx cluster.GrainContext)
	Start(*Empty, cluster.GrainContext) (*Empty, error)
	Describe(*DescribeAdminRequest, cluster.GrainContext) (*DescribeAdminResponse, error)
	Shutdown(*ShutdownRequest, cluster.GrainContext) (*ShutdownResponse, error)
}

// AdminGrainClient holds the base data for the AdminGrain
type AdminGrainClient struct {
	Identity string
	cluster  *cluster.Cluster
}

// Start requests the execution on to the cluster with CallOptions
func (g *AdminGrainClient) Start(r *Empty, opts ...cluster.GrainCallOption) (*Empty, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 0, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Admin", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &Empty{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// Describe requests the execution on to the cluster with CallOptions
func (g *AdminGrainClient) Describe(r *DescribeAdminRequest, opts ...cluster.GrainCallOption) (*DescribeAdminResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 1, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Admin", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &DescribeAdminResponse{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// Shutdown requests the execution on to the cluster with CallOptions
func (g *AdminGrainClient) Shutdown(r *ShutdownRequest, opts ...cluster.GrainCallOption) (*ShutdownResponse, error) {
	bytes, err := proto.Marshal(r)
	if err != nil {
		return nil, err
	}
	reqMsg := &cluster.GrainRequest{MethodIndex: 2, MessageData: bytes}
	resp, err := g.cluster.Call(g.Identity, "Admin", reqMsg, opts...)
	if err != nil {
		return nil, err
	}
	switch msg := resp.(type) {
	case *cluster.GrainResponse:
		result := &ShutdownResponse{}
		err = proto.Unmarshal(msg.MessageData, result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case *cluster.GrainErrorResponse:
		return nil, errors.New(msg.Err)
	default:
		return nil, errors.New("unknown response")
	}
}

// AdminActor represents the actor structure
type AdminActor struct {
	ctx     cluster.GrainContext
	inner   Admin
	Timeout time.Duration
}

// Receive ensures the lifecycle of the actor for the received message
func (a *AdminActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started: //pass
	case *cluster.ClusterInit:
		a.ctx = cluster.NewGrainContext(ctx, msg.Identity, msg.Cluster)
		a.inner = xAdminFactory()
		a.inner.Init(a.ctx)

		if a.Timeout > 0 {
			ctx.SetReceiveTimeout(a.Timeout)
		}
	case *actor.ReceiveTimeout:
		ctx.Poison(ctx.Self())
	case *actor.Stopped:
		a.inner.Terminate(a.ctx)
	case actor.AutoReceiveMessage: // pass
	case actor.SystemMessage: // pass

	case *cluster.GrainRequest:
		switch msg.MethodIndex {
		case 0:
			req := &Empty{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("Start(Empty) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.Start(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("Start(Empty) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)
		case 1:
			req := &DescribeAdminRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("Describe(DescribeAdminRequest) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.Describe(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("Describe(DescribeAdminRequest) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)
		case 2:
			req := &ShutdownRequest{}
			err := proto.Unmarshal(msg.MessageData, req)
			if err != nil {
				plog.Error("Shutdown(ShutdownRequest) proto.Unmarshal failed.", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			r0, err := a.inner.Shutdown(req, a.ctx)
			if err != nil {
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			bytes, err := proto.Marshal(r0)
			if err != nil {
				plog.Error("Shutdown(ShutdownRequest) proto.Marshal failed", logmod.Error(err))
				resp := &cluster.GrainErrorResponse{Err: err.Error()}
				ctx.Respond(resp)
				return
			}
			resp := &cluster.GrainResponse{MessageData: bytes}
			ctx.Respond(resp)

		}
	default:
		a.inner.ReceiveDefault(a.ctx)
	}
}
