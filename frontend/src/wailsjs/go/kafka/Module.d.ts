// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {kafka} from '../models';

export function Brokers(arg1:string):Promise<kafka.TabBrokers>;

export function Close():void;

export function Connect(arg1:string):Promise<kafka.State>;

export function Consumers(arg1:string):Promise<kafka.TabConsumers>;

export function CreateNewProject(arg1:string):Promise<kafka.State>;

export function DeleteProject(arg1:string):Promise<kafka.State>;

export function SaveAddress(arg1:string,arg2:string):Promise<kafka.State>;

export function SaveAuthMethod(arg1:string,arg2:string):Promise<kafka.State>;

export function SaveAuthPassword(arg1:string,arg2:string):Promise<kafka.State>;

export function SaveAuthUsername(arg1:string,arg2:string):Promise<kafka.State>;

export function SaveCurrentTab(arg1:string,arg2:string):Promise<kafka.State>;

export function StartTopicConsuming(arg1:string,arg2:string,arg3:number):Promise<kafka.ConsumeTopicOutput>;

export function State():Promise<kafka.State>;

export function StopTopicConsuming(arg1:string):Promise<Error>;

export function Topics(arg1:string):Promise<kafka.TabTopics>;
