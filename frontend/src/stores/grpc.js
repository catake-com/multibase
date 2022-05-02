import { defineStore } from "pinia";

import { RefreshProtoDescriptors } from "../wailsjs/go/main/App";

export const useGRPCStore = defineStore({
  id: "grpc",
  state: () => ({
    address: "",
    request: "",
    response: "",
    importPathList: [],
    protoFileList: [],
    nodes: [],
  }),
  actions: {
    setAddress(address) {
      this.address = address;
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

    addProtoFile(protoFile) {
      if (this.protoFileList.includes(protoFile)) {
        return;
      }

      this.protoFileList.push(protoFile);
      this.saveState();
    },

    removeProtoFile(protoFile) {
      this.protoFileList = this.protoFileList.filter((item) => item !== protoFile);

      this.saveState();
    },

    saveState() {
      const state = { address: this.address, importPathList: this.importPathList, protoFileList: this.protoFileList };

      localStorage.setItem("grpcState", JSON.stringify(state));

      this.refreshProtoDescriptors();
    },

    loadState() {
      const state = JSON.parse(localStorage.getItem("grpcState")) || {};

      if ("address" in state) {
        this.address = state.address;
      }

      if ("importPathList" in state) {
        this.importPathList = state.importPathList;
      }

      if ("protoFileList" in state) {
        this.protoFileList = state.protoFileList;
      }

      this.refreshProtoDescriptors();
    },

    refreshProtoDescriptors() {
      RefreshProtoDescriptors(this.importPathList, this.protoFileList)
        .then((nodes) => {
          this.nodes = nodes;
        })
        .catch((reason) => {
          this.response = reason;
        });
    },
  },
});
