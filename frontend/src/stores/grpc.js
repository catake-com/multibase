import { defineStore } from "pinia";

import {
  CreateNewProject,
  CreateNewForm,
  DeleteAllProtoFiles,
  RemoveForm,
  SelectMethod,
  SendRequest,
  StopRequest,
  OpenImportPath,
  RemoveImportPath,
  OpenProtoFile,
  SaveCurrentFormID,
  SaveAddress,
  SaveSplitterWidth,
  SaveRequestPayload,
  Project,
  DeleteProject,
  AddHeader,
  SaveHeaders,
  DeleteHeader,
  ReflectProto,
} from "../wailsjs/go/grpc/Module";
import { grpc } from "../wailsjs/go/models";

export const useGRPCStore = defineStore({
  id: "grpc",
  state: () => new grpc.Project(), // TODO: make object from class
  actions: {
    async createNewProject(projectID) {
      try {
        this.$state = await CreateNewProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async deleteProject(projectID) {
      try {
        this.$state = await DeleteProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async createNewForm(projectID) {
      try {
        this.$state = await CreateNewForm(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async removeForm(projectID, formID) {
      try {
        this.$state = await RemoveForm(projectID, formID);
      } catch (error) {
        console.log(error);
      }
    },

    async selectMethod(projectID, formID, methodID) {
      try {
        this.$state = await SelectMethod(projectID, formID, methodID);
      } catch (error) {
        this.forms[formID].response = error;
      }
    },

    async reflectProto(projectID, formID) {
      try {
        this.$state = await ReflectProto(projectID, formID, this.forms[formID].address);
      } catch (error) {
        this.forms[formID].response = error;
      }
    },

    async sendRequest(projectID, formID) {
      if (this.forms[formID].requestInProgress) {
        return;
      }

      this.forms[formID].requestInProgress = true;

      try {
        this.$state = await SendRequest(projectID, formID, this.forms[formID].address, this.forms[formID].request);
        this.forms[formID].requestInProgress = false;
      } catch (error) {
        this.forms[formID].requestInProgress = false;
        this.forms[formID].response = error;
      }
    },

    async stopRequest(projectID, formID) {
      if (!this.forms[formID].requestInProgress) {
        return;
      }

      try {
        this.$state = await StopRequest(projectID, formID);
        this.forms[formID].requestInProgress = false;
      } catch (error) {
        this.forms[formID].requestInProgress = false;
        this.forms[formID].response = error;
      }
    },

    async openImportPath(projectID) {
      try {
        this.$state = await OpenImportPath(projectID);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async removeImportPath(projectID, importPath) {
      try {
        this.$state = await RemoveImportPath(projectID, importPath);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async openProtoFile(projectID) {
      try {
        this.$state = await OpenProtoFile(projectID);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async deleteAllProtoFiles(projectID) {
      try {
        this.$state = await DeleteAllProtoFiles(projectID);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async saveCurrentFormID(projectID, currentFormID) {
      try {
        this.$state = await SaveCurrentFormID(projectID, currentFormID);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async saveAddress(projectID, formID, address) {
      try {
        this.$state = await SaveAddress(projectID, formID, address);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async addHeader(projectID, formID) {
      try {
        this.$state = await AddHeader(projectID, formID);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async saveHeaders(projectID, formID, headers) {
      try {
        this.$state = await SaveHeaders(projectID, formID, headers);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async deleteHeader(projectID, formID, headerID) {
      try {
        this.$state = await DeleteHeader(projectID, formID, headerID);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async saveSplitterWidth(projectID, splitterWidth) {
      try {
        this.$state = await SaveSplitterWidth(projectID, splitterWidth);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async saveRequestPayload(projectID, formID, requestPayload) {
      try {
        this.$state = await SaveRequestPayload(projectID, formID, requestPayload);
      } catch (error) {
        this.forms[this.currentFormID].response = error;
      }
    },

    async loadProject(projectID) {
      try {
        this.$state = await Project(projectID);
      } catch (error) {
        console.log(error);
      }
    },
  },
});
