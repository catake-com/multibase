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
  SaveSplitterWidth,
  SaveRequestPayload,
  State,
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
            request: "",
            response: "",
            requestInProgress: false,
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
      return CreateNewProject(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    async createNewForm(projectID) {
      return CreateNewForm(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    async removeForm(projectID, formID) {
      return RemoveForm(projectID, formID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    async selectFunction(projectID, formID, methodID) {
      return SelectFunction(projectID, formID, methodID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[formID].response = reason;
        });
    },

    async sendRequest(projectID, formID) {
      if (this.projects[projectID].forms[formID].requestInProgress) {
        return;
      }

      this.projects[projectID].forms[formID].requestInProgress = true;

      return SendRequest(
        projectID,
        formID,
        this.projects[projectID].forms[formID].address,
        this.projects[projectID].forms[formID].request
      )
        .then((state) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.projects[projectID].forms[formID].response = reason;
        });
    },

    async stopRequest(projectID, formID) {
      if (!this.projects[projectID].forms[formID].requestInProgress) {
        return;
      }

      return StopRequest(projectID, formID)
        .then((state) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.projects[projectID].forms[formID].response = reason;
        });
    },

    async openFilePath(projectID) {
      return OpenFilePath(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    async saveCurrentFormID(projectID, currentFormID) {
      return SaveCurrentFormID(projectID, currentFormID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    async saveAddress(projectID, formID, address) {
      return SaveAddress(projectID, formID, address)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    async saveSplitterWidth(projectID, address) {
      return SaveSplitterWidth(projectID, address)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    async saveRequestPayload(projectID, formID, requestPayload) {
      return SaveRequestPayload(projectID, formID, requestPayload)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    async loadState() {
      return State()
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },
  },
});
