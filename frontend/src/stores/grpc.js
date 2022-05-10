import { defineStore } from "pinia";

import { RefreshProtoDescriptors, SelectMethod, SendRequest } from "../wailsjs/go/main/App";

export const useGRPCStore = defineStore({
  id: "grpc",
  state: () => ({
    forms: {
      1: {
        address: "0.0.0.0:50051",
        selectedMethodID: "",
        request: "",
        response: "",
      },
    },
    currentFormID: 1,
    importPathList: [],
    protoFileList: [],
    nodes: [],
  }),
  actions: {
    createNewForm() {
      const formID = Math.floor(Date.now() * Math.random());
      this.forms[formID] = {
        address: "0.0.0.0:50051",
        selectedMethodID: "",
        request: "",
        response: "",
      };
      this.currentFormID = formID;

      this.saveState();
    },

    removeForm(formID) {
      if (Object.keys(this.forms).length <= 1) {
        return;
      }

      this.currentFormID = parseInt(Object.keys(this.forms)[0]);
      delete this.forms[formID];

      this.saveState();
    },

    selectMethod(formID, methodID) {
      SelectMethod(methodID)
        .then((payload) => {
          this.forms[formID].request = payload;
          this.forms[formID].selectedMethodID = methodID;
        })
        .catch((reason) => {
          this.forms[formID].response = reason;
        });

      this.saveState();
    },

    sendRequest(formID) {
      SendRequest(this.forms[formID].address, this.forms[formID].selectedMethodID, this.forms[formID].request)
        .then((response) => {
          this.forms[formID].response = response;
        })
        .catch((reason) => {
          this.forms[formID].response = reason;
        });

      this.saveState();
    },

    addImportPath(importPath) {
      if (this.importPathList.includes(importPath)) {
        return;
      }
      this.importPathList.push(importPath);

      this.saveState();
    },

    removeImportPath(importPath) {
      this.importPathList = this.importPathList.filter((item) => item !== importPath);

      this.saveState();
    },

    addProtoFile(protoFile, currentDir) {
      if (this.protoFileList.includes(protoFile)) {
        return;
      }

      const importPathList = [...this.importPathList];
      if (importPathList.length === 0) {
        importPathList.push(currentDir);
      }

      RefreshProtoDescriptors(importPathList, [protoFile, ...this.protoFileList])
        .then((nodes) => {
          this.nodes = nodes;

          this.importPathList = importPathList;
          this.protoFileList.push(protoFile);
          this.saveState();
        })
        .catch((reason) => {
          this.forms[this.currentFormID].response = reason;
        });
    },

    removeProtoFile(protoFile) {
      this.protoFileList = this.protoFileList.filter((item) => item !== protoFile);

      this.saveState();
    },

    saveState() {
      const state = {
        forms: this.forms,
        currentFormID: this.currentFormID,
        importPathList: this.importPathList,
        protoFileList: this.protoFileList,
      };

      localStorage.setItem("grpcState", JSON.stringify(state));
    },

    loadState() {
      const state = JSON.parse(localStorage.getItem("grpcState")) || {};

      if ("forms" in state) {
        this.forms = state.forms;
      }

      if ("currentFormID" in state) {
        this.currentFormID = state.currentFormID;
      }

      if ("importPathList" in state) {
        this.importPathList = state.importPathList;
      }

      if ("protoFileList" in state) {
        this.protoFileList = state.protoFileList;
      }

      this.loadNodes();
    },

    clearState() {
      localStorage.removeItem("grpcState");

      this.loadNodes();
    },

    loadNodes() {
      RefreshProtoDescriptors(this.importPathList, this.protoFileList)
        .then((nodes) => {
          this.nodes = nodes;
        })
        .catch((reason) => {
          this.forms[this.currentFormID].response = reason;
        });
    },
  },
});
