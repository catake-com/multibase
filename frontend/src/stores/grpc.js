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
  State,
} from "../wailsjs/go/grpc/Module";

export const useGRPCStore = defineStore({
  id: "grpc",
  state: () => ({
    projects: {
      "ae3d1fa3-09c7-4af0-a57f-65c24cbdf5f3": {
        forms: {
          "b7ce6ea8-c5f1-477f-bdb1-43814c2106ed": {
            address: "0.0.0.0:50051",
            selectedMethodID: "",
            request: "",
            response: "",
            requestInProgress: false,
          },
        },
        currentFormID: "b7ce6ea8-c5f1-477f-bdb1-43814c2106ed",
        importPathList: [],
        protoFileList: [],
        nodes: [],
      },
    },
  }),
  actions: {
    createNewProject(projectID) {
      CreateNewProject(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    createNewForm(projectID) {
      CreateNewForm(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    removeForm(projectID, formID) {
      RemoveForm(projectID, formID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    selectMethod(projectID, formID, methodID) {
      SelectMethod(projectID, formID, methodID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[formID].response = reason;
        });
    },

    sendRequest(projectID, formID) {
      if (this.projects[projectID].forms[formID].requestInProgress) {
        return;
      }

      this.projects[projectID].forms[formID].requestInProgress = true;

      SendRequest(
        projectID,
        formID,
        this.projects[projectID].forms[formID].address,
        this.projects[projectID].forms[formID].selectedMethodID,
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

    stopRequest(projectID, formID) {
      if (!this.projects[projectID].forms[formID].requestInProgress) {
        return;
      }

      StopRequest(projectID, formID)
        .then((state) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.projects[projectID].forms[formID].response = reason;
        });
    },

    openImportPath(projectID) {
      OpenImportPath(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    removeImportPath(projectID, importPath) {
      RemoveImportPath(projectID, importPath)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    openProtoFile(projectID) {
      OpenProtoFile(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    deleteAllProtoFiles(projectID) {
      DeleteAllProtoFiles(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    loadState() {
      State()
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },
  },
});
