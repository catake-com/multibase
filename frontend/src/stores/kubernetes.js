import { acceptHMRUpdate, defineStore } from "pinia";

import {
  CreateNewProject,
  DeleteProject,
  ProjectState,
  Connect,
  OverviewData,
  SelectNamespace,
  SaveCurrentTab,
} from "../wailsjs/go/kubernetes/Module";

export const useKubernetesStore = defineStore({
  id: "kubernetes",
  state: () => ({
    projectStates: {},
    overviewDataByProjectID: {},
  }),
  getters: {
    projectState: (state) => {
      return (projectID) => state.projectStates[projectID] || {};
    },
    overviewData: (state) => {
      return (projectID) => state.overviewDataByProjectID[projectID] || {};
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

    async loadOverviewData(projectID) {
      this.overviewDataByProjectID[projectID] = await OverviewData(projectID);
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
