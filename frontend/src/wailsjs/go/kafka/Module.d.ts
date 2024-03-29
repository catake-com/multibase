// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {kafka} from '../models';

export function Brokers(arg1:string):Promise<any>;

export function Close():Promise<void>;

export function Connect(arg1:string):Promise<any>;

export function Consumers(arg1:string):Promise<any>;

export function CreateNewProject(arg1:string):Promise<any>;

export function DeleteProject(arg1:string):Promise<void>;

export function ProjectState(arg1:string):Promise<any>;

export function SaveState(arg1:string,arg2:any):Promise<any>;

export function StartTopicConsuming(arg1:string,arg2:kafka.TopicConsumingStrategy,arg3:string,arg4:string,arg5:number):Promise<any>;

export function StopTopicConsuming(arg1:string):Promise<void>;

export function Topics(arg1:string):Promise<any>;
