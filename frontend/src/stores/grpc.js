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
  State,
  DeleteProject,
} from "../wailsjs/go/grpc/Module";

export const useGRPCStore = defineStore({
  id: "grpc",
  state: () => ({
    projects: {
      "ae3d1fa3-09c7-4af0-a57f-65c24cbdf5f3": {
        splitterWidth: 30,
        forms: {
          "b7ce6ea8-c5f1-477f-bdb1-43814c2106ed": {
            address: "0.0.0.0:50051",
            selectedMethodID: "",
            request: "",
            response: "",
            requestInProgress: false,
          },
        },
        formIDs: [],
        currentFormID: "b7ce6ea8-c5f1-477f-bdb1-43814c2106ed",
        importPathList: [],
        protoFileList: [],
        nodes: [],
      },
    },
  }),
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
        this.projects[projectID].forms[formID].response = error;
      }
    },

    async sendRequest(projectID, formID) {
      if (this.projects[projectID].forms[formID].requestInProgress) {
        return;
      }

      this.projects[projectID].forms[formID].requestInProgress = true;

      try {
        this.$state = await SendRequest(
          projectID,
          formID,
          this.projects[projectID].forms[formID].address,
          this.projects[projectID].forms[formID].request
        );
        this.projects[projectID].forms[formID].requestInProgress = false;
      } catch (error) {
        this.projects[projectID].forms[formID].requestInProgress = false;
        this.projects[projectID].forms[formID].response = error;
      }
    },

    async stopRequest(projectID, formID) {
      if (!this.projects[projectID].forms[formID].requestInProgress) {
        return;
      }

      try {
        this.$state = await StopRequest(projectID, formID);
        this.projects[projectID].forms[formID].requestInProgress = false;
      } catch (error) {
        this.projects[projectID].forms[formID].requestInProgress = false;
        this.projects[projectID].forms[formID].response = error;
      }
    },

    async openImportPath(projectID) {
      try {
        this.$state = await OpenImportPath(projectID);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async removeImportPath(projectID, importPath) {
      try {
        this.$state = await RemoveImportPath(projectID, importPath);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async openProtoFile(projectID) {
      try {
        this.$state = await OpenProtoFile(projectID);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async deleteAllProtoFiles(projectID) {
      try {
        this.$state = await DeleteAllProtoFiles(projectID);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveCurrentFormID(projectID, currentFormID) {
      try {
        this.$state = await SaveCurrentFormID(projectID, currentFormID);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveAddress(projectID, formID, address) {
      try {
        this.$state = await SaveAddress(projectID, formID, address);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveSplitterWidth(projectID, splitterWidth) {
      try {
        this.$state = await SaveSplitterWidth(projectID, splitterWidth);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveRequestPayload(projectID, formID, requestPayload) {
      try {
        this.$state = await SaveRequestPayload(projectID, formID, requestPayload);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async loadState() {
      try {
        this.$state = await State();
      } catch (error) {
        console.log(error);
      }
    },
  },
});
