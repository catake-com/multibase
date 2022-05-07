import { defineStore } from "pinia";

import { RefreshProtoDescriptors } from "../wailsjs/go/main/App";

export const useGRPCStore = defineStore({
  id: "grpc",
  state: () => ({
    forms: {
      1: {
        address: "0.0.0.0:50051",
        request: "",
        response: "",
      },
    },
    currentFormID: 1,
    importPathList: [],
    protoFileList: [],
    nodes: [],
  }),
  getters: {
    addressByFormID: (state) => {
      return (formID) => state.forms[formID].address;
    },

    requestByFormID: (state) => {
      return (formID) => state.forms[formID].request;
    },

    responseByFormID: (state) => {
      return (formID) => state.forms[formID].response;
    },
  },
  actions: {
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

      this.refreshProtoDescriptors();
    },

    clearState() {
      localStorage.removeItem("grpcState");

      this.refreshProtoDescriptors();
    },

    refreshProtoDescriptors() {
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
