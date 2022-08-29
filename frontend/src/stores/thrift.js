import { defineStore } from "pinia";

import {
  CreateNewProject,
  CreateNewForm,
  RemoveForm,
  SelectFunction,
  SendRequest,
  StopRequest,
  OpenFilePath,
  SaveCurrentFormID,
  SaveAddress,
  SaveIsMultiplexed,
  SaveSplitterWidth,
  SaveRequestPayload,
  State,
  AddHeader,
  SaveHeaders,
  DeleteHeader,
  DeleteProject,
} from "../wailsjs/go/thrift/Module";

export const useThriftStore = defineStore({
  id: "thrift",
  state: () => ({
    projects: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {
        splitterWidth: 30,
        forms: {
          "aba7bb0d-77f5-404c-a293-e133975ea67d": {
            address: "0.0.0.0:9090",
            selectedFunctionID: "",
            isMultiplexed: false,
            request: "",
            response: "",
            requestInProgress: false,
            headers: [{ id: "dd11abd9-80f5-494a-b85e-358c3103704a", key: "", value: "" }],
          },
        },
        formIDs: [],
        currentFormID: "aba7bb0d-77f5-404c-a293-e133975ea67d",
        filePath: [],
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

    async selectFunction(projectID, formID, methodID) {
      try {
        this.$state = await SelectFunction(projectID, formID, methodID);
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

    async openFilePath(projectID) {
      try {
        this.$state = await OpenFilePath(projectID);
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

    async saveIsMultiplexed(projectID, formID, isMultiplexed) {
      try {
        this.$state = await SaveIsMultiplexed(projectID, formID, isMultiplexed);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async addHeader(projectID, formID) {
      try {
        this.$state = await AddHeader(projectID, formID);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async saveHeaders(projectID, formID, headers) {
      try {
        this.$state = await SaveHeaders(projectID, formID, headers);
      } catch (error) {
        this.projects[projectID].forms[this.projects[projectID].currentFormID].response = error;
      }
    },

    async deleteHeader(projectID, formID, headerID) {
      try {
        this.$state = await DeleteHeader(projectID, formID, headerID);
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
