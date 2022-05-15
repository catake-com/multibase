import { defineStore } from "pinia";

export const useProjectStore = defineStore({
  id: "project",
  state: () => ({
    projects: {
      0: {
        type: "new",
      },
      1: {
        type: "grpc",
      },
    },
    openedTabs: [0, 1],
    currentProjectID: 0,
  }),
  actions: {
    createNewProject() {
      const projectID = Math.floor(Date.now() * Math.random());
      this.projects[projectID] = {
        type: "new",
      };
      this.openedTabs.push(projectID);
      this.currentProjectID = projectID;

      this.saveState();
    },

    closeProjectTab(projectID) {
      if (this.openedTabs.length <= 1) {
        return;
      }

      this.currentProjectID = this.openedTabs[0];

      if (this.projects[projectID].type === "new") {
        delete this.projects[projectID];
      }

      this.openedTabs = this.openedTabs.filter((pID) => pID !== parseInt(projectID));

      this.saveState();
    },

    saveState() {
      const state = {
        projects: this.projects,
        openedTabs: this.openedTabs,
        currentProjectID: this.currentProjectID,
      };

      localStorage.setItem("projectState", JSON.stringify(state));
    },

    loadState() {
      const state = JSON.parse(localStorage.getItem("projectState")) || {};

      if ("projects" in state) {
        this.projects = state.projects;
      }

      if ("openedTabs" in state) {
        this.openedTabs = state.openedTabs;
      }

      if ("currentProjectID" in state) {
        this.currentProjectID = state.currentProjectID;
      }
    },

    clearState() {
      localStorage.removeItem("projectState");

      this.loadNodes();
    },
  },
});
