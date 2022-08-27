// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {thrift} from '../models';

export function AddHeader(arg1:string,arg2:string):Promise<thrift.State>;

export function CreateNewForm(arg1:string):Promise<thrift.State>;

export function CreateNewProject(arg1:string):Promise<thrift.State>;

export function DeleteHeader(arg1:string,arg2:string,arg3:string):Promise<thrift.State>;

export function DeleteProject(arg1:string):Promise<thrift.State>;

export function OpenFilePath(arg1:string):Promise<thrift.State>;

export function RemoveForm(arg1:string,arg2:string):Promise<thrift.State>;

export function SaveAddress(arg1:string,arg2:string,arg3:string):Promise<thrift.State>;

export function SaveCurrentFormID(arg1:string,arg2:string):Promise<thrift.State>;

export function SaveHeaders(arg1:string,arg2:string,arg3:Array<thrift.StateProjectFormHeader>):Promise<thrift.State>;

export function SaveRequestPayload(arg1:string,arg2:string,arg3:string):Promise<thrift.State>;

export function SaveSplitterWidth(arg1:string,arg2:number):Promise<thrift.State>;

export function SaveState():Promise<Error>;

export function SelectFunction(arg1:string,arg2:string,arg3:string):Promise<thrift.State>;

export function SendRequest(arg1:string,arg2:string,arg3:string,arg4:string):Promise<thrift.State>;

export function State():Promise<thrift.State>;

export function StopRequest(arg1:string,arg2:string):Promise<thrift.State>;
