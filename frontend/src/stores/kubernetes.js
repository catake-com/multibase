import { acceptHMRUpdate, defineStore } from "pinia";

import {
  CreateNewProject,
  DeleteProject,
  ProjectState,
  Connect,
  OverviewData,
  SelectNamespace,
  SaveCurrentTab,
  Namespaces,
  WorkloadsPodsData,
  StartPortForwarding,
  StopPortForwarding,
} from "../wailsjs/go/kubernetes/Module";

export const useKubernetesStore = defineStore({
  id: "kubernetes",
  state: () => ({
    projectStates: {},
    overviewDataByProjectID: {},
    namespacesByProjectID: {},
    workloadsPodsDataByProjectID: {},
  }),
  getters: {
    projectState: (state) => {
      return (projectID) => state.projectStates[projectID] || {};
    },
    overviewData: (state) => {
      return (projectID) => state.overviewDataByProjectID[projectID] || {};
    },
    namespaces: (state) => {
      return (projectID) => state.namespacesByProjectID[projectID] || {};
    },
    workloadsPodsData: (state) => {
      return (projectID) => state.workloadsPodsDataByProjectID[projectID] || {};
    },
  },
  actions: {
    async createNewProject(projectID) {
      try {
        this.projectStates[projectID] = await CreateNewProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async deleteProject(projectID) {
      try {
        await DeleteProject(projectID);
        delete this.projects[projectID];
      } catch (error) {
        console.log(error);
      }
    },

    async saveCurrentTab(projectID, currentTab) {
      try {
        this.projectStates[projectID] = await SaveCurrentTab(projectID, currentTab);
      } catch (error) {
        console.log(error);
      }
    },

    async selectNamespace(projectID, selectedNamespace) {
      try {
        this.projectStates[projectID] = await SelectNamespace(projectID, selectedNamespace);
      } catch (error) {
        console.log(error);
      }
    },

    async startPortForwarding(projectID, namespace, pod, ports) {
      try {
        this.projectStates[projectID] = await StartPortForwarding(projectID, namespace, pod, ports);
      } catch (error) {
        console.log(error);
      }
    },

    async stopPortForwarding(projectID) {
      try {
        this.projectStates[projectID] = await StopPortForwarding(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async loadOverviewData(projectID) {
      this.overviewDataByProjectID[projectID] = await OverviewData(projectID);
    },

    async loadNamespaces(projectID) {
      this.namespacesByProjectID[projectID] = await Namespaces(projectID);
    },

    async loadWorkloadsPodsData(projectID) {
      this.workloadsPodsDataByProjectID[projectID] = await WorkloadsPodsData(projectID);
    },

    async connect(projectID, selectedCluster) {
      this.projectStates[projectID] = await Connect(projectID, selectedCluster);
    },

    async loadProject(projectID) {
      if (this.projectStates[projectID]) {
        return this.projectStates[projectID];
      }

      try {
        this.projectStates[projectID] = await ProjectState(projectID);
      } catch (error) {
        console.log(error);
      }
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useKubernetesStore, import.meta.hot));
}
