/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"fmt"

	"configcenter/src/common/metadata"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/condition"
	frtypes "configcenter/src/common/mapstr"
	"configcenter/src/scene_server/topo_server/core/inst"
	"configcenter/src/scene_server/topo_server/core/types"
)

// CreateInst create a new inst
func (s *topoService) CreateInst(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {

	// /inst/{owner_id}/{obj_id}

	objID := pathParams("obj_id")

	obj, err := s.core.ObjectOperation().FindSingleObject(params, objID)
	if nil != err {
		blog.Errorf("failed to search the inst, %s", err.Error())
		return nil, err
	}

	setInst, err := s.core.InstOperation().CreateInst(params, obj, data)
	if nil != err {
		blog.Errorf("failed to create a new %s, %s", objID, err.Error())
		return nil, err
	}

	return setInst.ToMapStr(), nil
}

// DeleteInst delete the inst
func (s *topoService) DeleteInst(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {

	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(pathParams("obj_id"))

	objs, err := s.core.ObjectOperation().FindObject(params, cond)
	if nil != err {
		blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
		return nil, err
	}

	innerCond := condition.CreateCondition()
	paramPath := frtypes.MapStr{}
	paramPath.Set("inst_id", pathParams("inst_id"))
	id, err := paramPath.Int64("inst_id")
	if nil != err {
		blog.Errorf("[api-inst] failed to parse the path params id(%s), error info is %s ", pathParams("inst_id"), err.Error())
		return nil, err
	}
	innerCond.Field(common.BKInstIDField).Eq(id)
	for _, objItem := range objs {
		err = s.core.InstOperation().DeleteInst(params, objItem, innerCond)
		if nil != err {
			blog.Errorf("[api-inst] failed to delete the object(%s) inst (%s), error info is %s", objItem.GetID(), pathParams("inst_id"), err.Error())
			return nil, err
		}
	}

	return nil, err
}

// UpdateInst update the inst
func (s *topoService) UpdateInst(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {

	// /inst/{owner_id}/{obj_id}/{inst_id}

	objID := pathParams("obj_id")

	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(objID)

	objs, err := s.core.ObjectOperation().FindObject(params, cond)
	if nil != err {
		blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
		return nil, err
	}

	innerCond := condition.CreateCondition()
	paramPath := frtypes.MapStr{}
	paramPath.Set("inst_id", pathParams("inst_id"))
	id, err := paramPath.Int64("inst_id")
	if nil != err {
		blog.Errorf("[api-inst] failed to parse the path params id(%s), error info is %s ", pathParams("inst_id"), err.Error())
		return nil, err
	}
	innerCond.Field(common.BKInstIDField).Eq(id)
	for _, objItem := range objs {
		err = s.core.InstOperation().UpdateInst(params, data, objItem, innerCond)
		if nil != err {
			blog.Errorf("[api-inst] failed to update the object(%s) inst (%s),the data (%#v), error info is %s", objItem.GetID(), pathParams("inst_id"), data, err.Error())
			return nil, err
		}
	}

	return nil, err
}

// SearchInst search the inst
func (s *topoService) SearchInst(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {
	fmt.Println("SearchInst")
	// /inst/search/{owner_id}/{obj_id}

	objID := pathParams("obj_id")

	// query the objects
	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(objID)

	objs, err := s.core.ObjectOperation().FindObject(params, cond)
	if nil != err {
		blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
		return nil, err
	}

	// construct the query inst condition
	count := 0
	instRst := make([]inst.Inst, 0)
	queryCond := &metadata.QueryInput{}

	if err := data.MarshalJSONInto(queryCond); nil != err {
		blog.Errorf("[api-inst] failed to parse the data and the condition, the input (%#v), error info is %s", data, err.Error())
		return nil, err
	}

	innerQueryCond, err := frtypes.NewFromInterface(queryCond.Condition)
	if nil != err {
		blog.Errorf("[api-inst] failed to parse the condition, %s", err.Error())
		return nil, err
	}

	if err := cond.Parse(innerQueryCond); nil != err {
		blog.Errorf("[api-inst] failed to parse the condition(%#v)", innerQueryCond)
		return nil, err
	}
	queryCond.Condition = cond.ToMapStr()

	fmt.Println("the query condition:", queryCond)

	// query insts
	for _, objItem := range objs {

		cnt, instItems, err := s.core.InstOperation().FindInst(params, objItem, queryCond)
		if nil != err {
			blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
			return nil, err
		}
		count = count + cnt
		instRst = append(instRst, instItems...)
	}

	result := frtypes.MapStr{}
	result.Set("count", count)
	result.Set("info", instRst)
	return result, nil
}

// SearchInstAndAssociationDetail search the inst with association details
func (s *topoService) SearchInstAndAssociationDetail(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {
	fmt.Println("SearchInstAndAssociationDetail")
	// /inst/search/owner/{owner_id}/object/{obj_id}/detail

	objID := pathParams("obj_id")

	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(objID)

	objs, err := s.core.ObjectOperation().FindObject(params, cond)
	if nil != err {
		blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
		return nil, err
	}

	count := 0
	instRst := make([]inst.Inst, 0)
	queryCond := &metadata.QueryInput{}
	if err := data.MarshalJSONInto(queryCond); nil != err {
		blog.Errorf("[api-inst] failed to parse the data and the condition, the input (%#v), error info is %s", data, err.Error())
		return nil, err
	}

	for _, objItem := range objs {

		cnt, instItems, err := s.core.InstOperation().FindInst(params, objItem, queryCond)
		if nil != err {
			blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
			return nil, err
		}
		count = count + cnt
		instRst = append(instRst, instItems...)
	}

	result := frtypes.MapStr{}
	result.Set("count", count)
	result.Set("info", instRst)

	return result, nil
}

// SearchInstByObject search the inst of the object
func (s *topoService) SearchInstByObject(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {

	// /inst/search/owner/{owner_id}/object/{obj_id}

	objID := pathParams("obj_id")

	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(objID)

	objs, err := s.core.ObjectOperation().FindObject(params, cond)
	if nil != err {
		blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
		return nil, err
	}

	count := 0
	instRst := make([]inst.Inst, 0)
	queryCond := &metadata.QueryInput{}
	if err := data.MarshalJSONInto(queryCond); nil != err {
		blog.Errorf("[api-inst] failed to parse the data and the condition, the input (%#v), error info is %s", data, err.Error())
		return nil, err
	}

	for _, objItem := range objs {

		cnt, instItems, err := s.core.InstOperation().FindInst(params, objItem, queryCond)
		if nil != err {
			blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
			return nil, err
		}
		count = count + cnt
		instRst = append(instRst, instItems...)
	}

	result := frtypes.MapStr{}
	result.Set("count", count)
	result.Set("info", instRst)

	return result, nil

}

// SearchInstByAssociation search inst by the association inst
func (s *topoService) SearchInstByAssociation(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {
	fmt.Println("SearchInstByAssociation")
	// /inst/association/search/owner/{owner_id}/object/{obj_id}

	objID := pathParams("obj_id")

	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(objID)

	objs, err := s.core.ObjectOperation().FindObject(params, cond)
	if nil != err {
		blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
		return nil, err
	}

	count := 0
	instRst := make([]inst.Inst, 0)
	queryCond := &metadata.QueryInput{}
	if err := data.MarshalJSONInto(queryCond); nil != err {
		blog.Errorf("[api-inst] failed to parse the data and the condition, the input (%#v), error info is %s", data, err.Error())
		return nil, err
	}

	for _, objItem := range objs {

		cnt, instItems, err := s.core.InstOperation().FindInst(params, objItem, queryCond)
		if nil != err {
			blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
			return nil, err
		}
		count = count + cnt
		instRst = append(instRst, instItems...)
	}

	result := frtypes.MapStr{}
	result.Set("count", count)
	result.Set("info", instRst)

	return result, nil
}

// SearchInstByInstID search the inst by inst ID
func (s *topoService) SearchInstByInstID(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {
	fmt.Println("SearchInstByInstID")
	// /inst/search/{owner_id}/{obj_id}/{inst_id}

	objID := pathParams("obj_id")

	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(objID)
	cond.Field(common.BKInstIDField).Eq(pathParams("inst_id"))

	objs, err := s.core.ObjectOperation().FindObject(params, cond)
	if nil != err {
		blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
		return nil, err
	}

	count := 0
	instRst := make([]inst.Inst, 0)
	queryCond := &metadata.QueryInput{}
	if err := data.MarshalJSONInto(queryCond); nil != err {
		blog.Errorf("[api-inst] failed to parse the data and the condition, the input (%#v), error info is %s", data, err.Error())
		return nil, err
	}

	for _, objItem := range objs {

		cnt, instItems, err := s.core.InstOperation().FindInst(params, objItem, queryCond)
		if nil != err {
			blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
			return nil, err
		}
		count = count + cnt
		instRst = append(instRst, instItems...)
	}

	result := frtypes.MapStr{}
	result.Set("count", count)
	result.Set("info", instRst)

	return result, nil
}

// SearchInstChildTopo search the child inst topo for a inst
func (s *topoService) SearchInstChildTopo(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {
	fmt.Println("SearchInstChildTopo")
	// /inst/search/topo/owner/{owner_id}/object/{object_id}/inst/{inst_id}

	objID := pathParams("object_id")

	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(objID)

	objs, err := s.core.ObjectOperation().FindObject(params, cond)
	if nil != err {
		blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
		return nil, err
	}

	data.Set(common.BKInstIDField, pathParams("inst_id"))

	count := 0
	instRst := make([]inst.Inst, 0)
	queryCond := &metadata.QueryInput{}

	if err := data.MarshalJSONInto(queryCond); nil != err {
		blog.Errorf("[api-inst] failed to parse the data and the condition, the input (%#v), error info is %s", data, err.Error())
		return nil, err
	}

	for _, objItem := range objs {

		cnt, instItems, err := s.core.InstOperation().FindInst(params, objItem, queryCond)
		if nil != err {
			blog.Errorf("[api-inst] failed to find the objects(%s), error info is %s", pathParams("obj_id"), err.Error())
			return nil, err
		}
		count = count + cnt
		instRst = append(instRst, instItems...)
	}

	result := frtypes.MapStr{}
	result.Set("count", count)
	result.Set("info", instRst)

	return result, nil

}

// SearchInstTopo search the inst topo
func (s *topoService) SearchInstTopo(params types.ContextParams, pathParams, queryParams ParamsGetter, data frtypes.MapStr) (interface{}, error) {
	fmt.Println("SearchInstTopo")
	// /inst/association/topo/search/owner/{owner_id}/object/{object_id}/inst/{inst_id}

	objID := pathParams("object_id")

	cond := condition.CreateCondition()
	cond.Field(common.BKOwnerIDField).Eq(params.SupplierAccount)
	cond.Field(common.BKObjIDField).Eq(objID)

	return nil, nil
}
