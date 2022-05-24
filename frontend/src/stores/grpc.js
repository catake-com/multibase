import { defineStore } from "pinia";

import { RefreshProtoDescriptors, SelectMethod, SendRequest, StopRequest } from "../wailsjs/go/grpc/Module";

export const useGRPCStore = defineStore({
  id: "grpc",
  state: () => ({
    projects: {
      1: {
        forms: {
          1: {
            address: "0.0.0.0:50051",
            selectedMethodID: "",
            request: "",
            response: "",
            requestInProgress: false,
          },
        },
        currentFormID: 1,
        importPathList: [],
        protoFileList: [],
        nodes: [],
      },
    },
  }),
  actions: {
    createNewProject(projectID) {
      this.projects[projectID] = {
        forms: {
          1: {
            address: "0.0.0.0:50051",
            selectedMethodID: "",
            request: "",
            response: "",
            requestInProgress: false,
          },
        },
        currentFormID: 1,
        importPathList: [],
        protoFileList: [],
        nodes: [],
      };

      this.saveState();
    },

    createNewForm(projectID) {
      const formID = Math.floor(Date.now() * Math.random());
      this.projects[projectID].forms[formID] = {
        address: "0.0.0.0:50051",
        selectedMethodID: "",
        request: "",
        response: "",
      };
      this.projects[projectID].currentFormID = formID;

      this.saveState();
    },

    removeForm(projectID, formID) {
      if (Object.keys(this.projects[projectID].forms).length <= 1) {
        return;
      }

      delete this.projects[projectID].forms[formID];
      this.projects[projectID].currentFormID = parseInt(Object.keys(this.projects[projectID].forms)[0]);

      this.saveState();
    },

    selectMethod(projectID, formID, methodID) {
      SelectMethod(projectID, methodID)
        .then((payload) => {
          this.projects[projectID].forms[formID].request = payload;
          this.projects[projectID].forms[formID].selectedMethodID = methodID;
        })
        .catch((reason) => {
          this.projects[projectID].forms[formID].response = reason;
        });

      this.saveState();
    },

    sendRequest(projectID, formID) {
      if (this.projects[projectID].forms[formID].requestInProgress) {
        return;
      }

      this.projects[projectID].forms[formID].requestInProgress = true;

      SendRequest(
        projectID,
        parseInt(formID),
        this.projects[projectID].forms[formID].address,
        this.projects[projectID].forms[formID].selectedMethodID,
        this.projects[projectID].forms[formID].request
      )
        .then((response) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.projects[projectID].forms[formID].response = response;
        })
        .catch((reason) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.projects[projectID].forms[formID].response = reason;
        });

      this.saveState();
    },

    stopRequest(projectID, formID) {
      if (!this.projects[projectID].forms[formID].requestInProgress) {
        return;
      }

      StopRequest(projectID, parseInt(formID))
        .then((response) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.projects[projectID].forms[formID].response = response;
        })
        .catch((reason) => {
          this.projects[projectID].forms[formID].requestInProgress = false;
          this.projects[projectID].forms[formID].response = reason;
        });

      this.saveState();
    },

    addImportPath(projectID, importPath) {
      if (this.projects[projectID].importPathList.includes(importPath)) {
        return;
      }
      this.projects[projectID].importPathList.push(importPath);

      this.saveState();
    },

    removeImportPath(projectID, importPath) {
      this.projects[projectID].importPathList = this.projects[projectID].importPathList.filter(
        (item) => item !== importPath
      );

      this.saveState();
    },

    addProtoFile(projectID, protoFile, currentDir) {
      if (this.projects[projectID].protoFileList.includes(protoFile)) {
        return;
      }

      const importPathList = [...this.projects[projectID].importPathList];
      if (importPathList.length === 0) {
        importPathList.push(currentDir);
      }

      RefreshProtoDescriptors(projectID, importPathList, [protoFile, ...this.projects[projectID].protoFileList])
        .then((nodes) => {
          this.projects[projectID].nodes = nodes;

          this.projects[projectID].importPathList = importPathList;
          this.projects[projectID].protoFileList.push(protoFile);
          this.saveState();
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.projects[projectID].currentFormID].response = reason;
        });
    },

    removeProtoFile(protoFile) {
      this.projects[projectID].protoFileList = this.projects[projectID].protoFileList.filter(
        (item) => item !== protoFile
      );

      this.saveState();
    },

    saveState() {
      const state = {
        projects: this.projects,
      };

      localStorage.setItem("grpcState", JSON.stringify(state));
    },

    loadState() {
      const state = JSON.parse(localStorage.getItem("grpcState")) || {};

      if ("projects" in state) {
        this.projects = state.projects;

        Object.entries(state.projects).forEach(([projectID, project]) => {
          if ("forms" in project) {
            Object.entries(project.forms).forEach(([_, form]) => (form.requestInProgress = false));
          }

          this.loadNodes(parseInt(projectID));
        });
      }
    },

    clearState() {
      localStorage.removeItem("grpcState");

      // TODO: reset state properly
      // this.loadNodes();
    },

    loadNodes(projectID) {
      RefreshProtoDescriptors(
        projectID,
        this.projects[projectID].importPathList,
        this.projects[projectID].protoFileList
      )
        .then((nodes) => {
          this.projects[projectID].nodes = nodes;
        })
        .catch((reason) => {
          this.projects[projectID].forms[this.currentFormID].response = reason;
        });
    },
  },
});
