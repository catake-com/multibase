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
  state: () => ({
    projects: {},
  }),
  getters: {
    project: (state) => {
      return (projectID) => state.projects[projectID];
    },
  },
  actions: {
    async createNewProject(projectID) {
      try {
        this.projects[projectID] = await CreateNewProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async deleteProject(projectID) {
      try {
        this.projects[projectID] = await DeleteProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async createNewForm(projectID) {
      try {
        this.projects[projectID] = await CreateNewForm(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async removeForm(projectID, formID) {
      try {
        this.projects[projectID] = await RemoveForm(projectID, formID);
      } catch (error) {
        console.log(error);
      }
    },

    async selectMethod(projectID, formID, methodID) {
      try {
        this.projects[projectID] = await SelectMethod(projectID, formID, methodID);
      } catch (error) {
        this.projectID[projectID].forms[formID].response = error;
      }
    },

    async reflectProto(projectID, formID) {
      try {
        this.projects[projectID] = await ReflectProto(
          projectID,
          formID,
          this.projectID[projectID].forms[formID].address
        );
      } catch (error) {
        this.projectID[projectID].forms[formID].response = error;
      }
    },

    async sendRequest(projectID, formID) {
      if (this.projectID[projectID].forms[formID].requestInProgress) {
        return;
      }

      this.projectID[projectID].forms[formID].requestInProgress = true;

      try {
        this.projects[projectID] = await SendRequest(
          projectID,
          formID,
          this.projectID[projectID].forms[formID].address,
          this.projectID[projectID].forms[formID].request
        );
        this.projectID[projectID].forms[formID].requestInProgress = false;
      } catch (error) {
        this.projectID[projectID].forms[formID].requestInProgress = false;
        this.projectID[projectID].forms[formID].response = error;
      }
    },

    async stopRequest(projectID, formID) {
      if (!this.projectID[projectID].forms[formID].requestInProgress) {
        return;
      }

      try {
        this.projects[projectID] = await StopRequest(projectID, formID);
        this.projectID[projectID].forms[formID].requestInProgress = false;
      } catch (error) {
        this.projectID[projectID].forms[formID].requestInProgress = false;
        this.projectID[projectID].forms[formID].response = error;
      }
    },

    async openImportPath(projectID) {
      try {
        this.projects[projectID] = await OpenImportPath(projectID);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async removeImportPath(projectID, importPath) {
      try {
        this.projects[projectID] = await RemoveImportPath(projectID, importPath);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async openProtoFile(projectID) {
      try {
        this.projects[projectID] = await OpenProtoFile(projectID);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async deleteAllProtoFiles(projectID) {
      try {
        this.projects[projectID] = await DeleteAllProtoFiles(projectID);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveCurrentFormID(projectID, currentFormID) {
      try {
        this.projects[projectID] = await SaveCurrentFormID(projectID, currentFormID);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveAddress(projectID, formID, address) {
      try {
        this.projects[projectID] = await SaveAddress(projectID, formID, address);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async addHeader(projectID, formID) {
      try {
        this.projects[projectID] = await AddHeader(projectID, formID);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveHeaders(projectID, formID, headers) {
      try {
        this.projects[projectID] = await SaveHeaders(projectID, formID, headers);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async deleteHeader(projectID, formID, headerID) {
      try {
        this.projects[projectID] = await DeleteHeader(projectID, formID, headerID);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveSplitterWidth(projectID, splitterWidth) {
      try {
        this.projects[projectID] = await SaveSplitterWidth(projectID, splitterWidth);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveRequestPayload(projectID, formID, requestPayload) {
      try {
        this.projects[projectID] = await SaveRequestPayload(projectID, formID, requestPayload);
      } catch (error) {
        this.projectID[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async loadProject(projectID) {
      if (this.projects[projectID]) {
        return this.projects[projectID];
      }

      try {
        this.projects[projectID] = await Project(projectID);
      } catch (error) {
        console.log(error);
      }
    },
  },
});
