// Copyright (c) Mondoo, Inc.
// SPDX-License-Identifier: BUSL-1.1

// Code generated by resources. DO NOT EDIT.

package resources

import (
	"errors"

	"go.mondoo.com/cnquery/llx"
	"go.mondoo.com/cnquery/providers-sdk/v1/plugin"
	"go.mondoo.com/cnquery/types"
)

var resourceFactories map[string]plugin.ResourceFactory

func init() {
	resourceFactories = map[string]plugin.ResourceFactory {
		"gitlab.group": {
			Init: initGitlabGroup,
			Create: createGitlabGroup,
		},
		"gitlab.project": {
			// to override args, implement: initGitlabProject(runtime *plugin.Runtime, args map[string]*llx.RawData) (map[string]*llx.RawData, plugin.Resource, error)
			Create: createGitlabProject,
		},
	}
}

// NewResource is used by the runtime of this plugin to create new resources.
// Its arguments may be provided by users. This function is generally not
// used by initializing resources from recordings or from lists.
func NewResource(runtime *plugin.Runtime, name string, args map[string]*llx.RawData) (plugin.Resource, error) {
	f, ok := resourceFactories[name]
	if !ok {
		return nil, errors.New("cannot find resource " + name + " in this provider")
	}

	if f.Init != nil {
		cargs, res, err := f.Init(runtime, args)
		if err != nil {
			return res, err
		}

		if res != nil {
			id := name+"\x00"+res.MqlID()
			if x, ok := runtime.Resources.Get(id); ok {
				return x, nil
			}
			runtime.Resources.Set(id, res)
			return res, nil
		}

		args = cargs
	}

	res, err := f.Create(runtime, args)
	if err != nil {
		return nil, err
	}

	id := name+"\x00"+res.MqlID()
	if x, ok := runtime.Resources.Get(id); ok {
		return x, nil
	}

	runtime.Resources.Set(id, res)
	return res, nil
}

// CreateResource is used by the runtime of this plugin to create resources.
// Its arguments must be complete and pre-processed. This method is used
// for initializing resources from recordings or from lists.
func CreateResource(runtime *plugin.Runtime, name string, args map[string]*llx.RawData) (plugin.Resource, error) {
	f, ok := resourceFactories[name]
	if !ok {
		return nil, errors.New("cannot find resource " + name + " in this provider")
	}

	res, err := f.Create(runtime, args)
	if err != nil {
		return nil, err
	}

	id := name+"\x00"+res.MqlID()
	if x, ok := runtime.Resources.Get(id); ok {
		return x, nil
	}

	runtime.Resources.Set(id, res)
	return res, nil
}

var getDataFields = map[string]func(r plugin.Resource) *plugin.DataRes{
	"gitlab.group.id": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetId()).ToDataRes(types.Int)
	},
	"gitlab.group.name": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetName()).ToDataRes(types.String)
	},
	"gitlab.group.path": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetPath()).ToDataRes(types.String)
	},
	"gitlab.group.description": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetDescription()).ToDataRes(types.String)
	},
	"gitlab.group.webURL": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetWebURL()).ToDataRes(types.String)
	},
	"gitlab.group.visibility": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetVisibility()).ToDataRes(types.String)
	},
	"gitlab.group.requireTwoFactorAuthentication": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetRequireTwoFactorAuthentication()).ToDataRes(types.Bool)
	},
	"gitlab.group.preventForkingOutsideGroup": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetPreventForkingOutsideGroup()).ToDataRes(types.Bool)
	},
	"gitlab.group.emailsDisabled": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetEmailsDisabled()).ToDataRes(types.Bool)
	},
	"gitlab.group.mentionsDisabled": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetMentionsDisabled()).ToDataRes(types.Bool)
	},
	"gitlab.group.projects": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabGroup).GetProjects()).ToDataRes(types.Array(types.Resource("gitlab.project")))
	},
	"gitlab.project.id": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabProject).GetId()).ToDataRes(types.Int)
	},
	"gitlab.project.name": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabProject).GetName()).ToDataRes(types.String)
	},
	"gitlab.project.path": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabProject).GetPath()).ToDataRes(types.String)
	},
	"gitlab.project.description": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabProject).GetDescription()).ToDataRes(types.String)
	},
	"gitlab.project.visibility": func(r plugin.Resource) *plugin.DataRes {
		return (r.(*mqlGitlabProject).GetVisibility()).ToDataRes(types.String)
	},
}

func GetData(resource plugin.Resource, field string, args map[string]*llx.RawData) *plugin.DataRes {
	f, ok := getDataFields[resource.MqlName()+"."+field]
	if !ok {
		return &plugin.DataRes{Error: "cannot find '" + field + "' in resource '" + resource.MqlName() + "'"}
	}

	return f(resource)
}

var setDataFields = map[string]func(r plugin.Resource, v *llx.RawData) bool {
	"gitlab.group.__id": func(r plugin.Resource, v *llx.RawData) (ok bool) {
			r.(*mqlGitlabGroup).__id, ok = v.Value.(string)
			return
		},
	"gitlab.group.id": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).Id, ok = plugin.RawToTValue[int64](v.Value, v.Error)
		return
	},
	"gitlab.group.name": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).Name, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
	"gitlab.group.path": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).Path, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
	"gitlab.group.description": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).Description, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
	"gitlab.group.webURL": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).WebURL, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
	"gitlab.group.visibility": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).Visibility, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
	"gitlab.group.requireTwoFactorAuthentication": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).RequireTwoFactorAuthentication, ok = plugin.RawToTValue[bool](v.Value, v.Error)
		return
	},
	"gitlab.group.preventForkingOutsideGroup": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).PreventForkingOutsideGroup, ok = plugin.RawToTValue[bool](v.Value, v.Error)
		return
	},
	"gitlab.group.emailsDisabled": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).EmailsDisabled, ok = plugin.RawToTValue[bool](v.Value, v.Error)
		return
	},
	"gitlab.group.mentionsDisabled": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).MentionsDisabled, ok = plugin.RawToTValue[bool](v.Value, v.Error)
		return
	},
	"gitlab.group.projects": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabGroup).Projects, ok = plugin.RawToTValue[[]interface{}](v.Value, v.Error)
		return
	},
	"gitlab.project.__id": func(r plugin.Resource, v *llx.RawData) (ok bool) {
			r.(*mqlGitlabProject).__id, ok = v.Value.(string)
			return
		},
	"gitlab.project.id": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabProject).Id, ok = plugin.RawToTValue[int64](v.Value, v.Error)
		return
	},
	"gitlab.project.name": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabProject).Name, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
	"gitlab.project.path": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabProject).Path, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
	"gitlab.project.description": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabProject).Description, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
	"gitlab.project.visibility": func(r plugin.Resource, v *llx.RawData) (ok bool) {
		r.(*mqlGitlabProject).Visibility, ok = plugin.RawToTValue[string](v.Value, v.Error)
		return
	},
}

func SetData(resource plugin.Resource, field string, val *llx.RawData) error {
	f, ok := setDataFields[resource.MqlName() + "." + field]
	if !ok {
		return errors.New("[gitlab] cannot set '"+field+"' in resource '"+resource.MqlName()+"', field not found")
	}

	if ok := f(resource, val); !ok {
		return errors.New("[gitlab] cannot set '"+field+"' in resource '"+resource.MqlName()+"', type does not match")
	}
	return nil
}

func SetAllData(resource plugin.Resource, args map[string]*llx.RawData) error {
	var err error
	for k, v := range args {
		if err = SetData(resource, k, v); err != nil {
			return err
		}
	}
	return nil
}

// mqlGitlabGroup for the gitlab.group resource
type mqlGitlabGroup struct {
	MqlRuntime *plugin.Runtime
	__id string
	// optional: if you define mqlGitlabGroupInternal it will be used here
	Id plugin.TValue[int64]
	Name plugin.TValue[string]
	Path plugin.TValue[string]
	Description plugin.TValue[string]
	WebURL plugin.TValue[string]
	Visibility plugin.TValue[string]
	RequireTwoFactorAuthentication plugin.TValue[bool]
	PreventForkingOutsideGroup plugin.TValue[bool]
	EmailsDisabled plugin.TValue[bool]
	MentionsDisabled plugin.TValue[bool]
	Projects plugin.TValue[[]interface{}]
}

// createGitlabGroup creates a new instance of this resource
func createGitlabGroup(runtime *plugin.Runtime, args map[string]*llx.RawData) (plugin.Resource, error) {
	res := &mqlGitlabGroup{
		MqlRuntime: runtime,
	}

	err := SetAllData(res, args)
	if err != nil {
		return res, err
	}

	if res.__id == "" {
	res.__id, err = res.id()
		if err != nil {
			return nil, err
		}
	}

	if runtime.HasRecording {
		args, err = runtime.ResourceFromRecording("gitlab.group", res.__id)
		if err != nil || args == nil {
			return res, err
		}
		return res, SetAllData(res, args)
	}

	return res, nil
}

func (c *mqlGitlabGroup) MqlName() string {
	return "gitlab.group"
}

func (c *mqlGitlabGroup) MqlID() string {
	return c.__id
}

func (c *mqlGitlabGroup) GetId() *plugin.TValue[int64] {
	return &c.Id
}

func (c *mqlGitlabGroup) GetName() *plugin.TValue[string] {
	return &c.Name
}

func (c *mqlGitlabGroup) GetPath() *plugin.TValue[string] {
	return &c.Path
}

func (c *mqlGitlabGroup) GetDescription() *plugin.TValue[string] {
	return &c.Description
}

func (c *mqlGitlabGroup) GetWebURL() *plugin.TValue[string] {
	return &c.WebURL
}

func (c *mqlGitlabGroup) GetVisibility() *plugin.TValue[string] {
	return &c.Visibility
}

func (c *mqlGitlabGroup) GetRequireTwoFactorAuthentication() *plugin.TValue[bool] {
	return &c.RequireTwoFactorAuthentication
}

func (c *mqlGitlabGroup) GetPreventForkingOutsideGroup() *plugin.TValue[bool] {
	return &c.PreventForkingOutsideGroup
}

func (c *mqlGitlabGroup) GetEmailsDisabled() *plugin.TValue[bool] {
	return &c.EmailsDisabled
}

func (c *mqlGitlabGroup) GetMentionsDisabled() *plugin.TValue[bool] {
	return &c.MentionsDisabled
}

func (c *mqlGitlabGroup) GetProjects() *plugin.TValue[[]interface{}] {
	return plugin.GetOrCompute[[]interface{}](&c.Projects, func() ([]interface{}, error) {
		if c.MqlRuntime.HasRecording {
			d, err := c.MqlRuntime.FieldResourceFromRecording("gitlab.group", c.__id, "projects")
			if err != nil {
				return nil, err
			}
			if d != nil {
				return d.Value.([]interface{}), nil
			}
		}

		return c.projects()
	})
}

// mqlGitlabProject for the gitlab.project resource
type mqlGitlabProject struct {
	MqlRuntime *plugin.Runtime
	__id string
	// optional: if you define mqlGitlabProjectInternal it will be used here
	Id plugin.TValue[int64]
	Name plugin.TValue[string]
	Path plugin.TValue[string]
	Description plugin.TValue[string]
	Visibility plugin.TValue[string]
}

// createGitlabProject creates a new instance of this resource
func createGitlabProject(runtime *plugin.Runtime, args map[string]*llx.RawData) (plugin.Resource, error) {
	res := &mqlGitlabProject{
		MqlRuntime: runtime,
	}

	err := SetAllData(res, args)
	if err != nil {
		return res, err
	}

	if res.__id == "" {
	res.__id, err = res.id()
		if err != nil {
			return nil, err
		}
	}

	if runtime.HasRecording {
		args, err = runtime.ResourceFromRecording("gitlab.project", res.__id)
		if err != nil || args == nil {
			return res, err
		}
		return res, SetAllData(res, args)
	}

	return res, nil
}

func (c *mqlGitlabProject) MqlName() string {
	return "gitlab.project"
}

func (c *mqlGitlabProject) MqlID() string {
	return c.__id
}

func (c *mqlGitlabProject) GetId() *plugin.TValue[int64] {
	return &c.Id
}

func (c *mqlGitlabProject) GetName() *plugin.TValue[string] {
	return &c.Name
}

func (c *mqlGitlabProject) GetPath() *plugin.TValue[string] {
	return &c.Path
}

func (c *mqlGitlabProject) GetDescription() *plugin.TValue[string] {
	return &c.Description
}

func (c *mqlGitlabProject) GetVisibility() *plugin.TValue[string] {
	return &c.Visibility
}
