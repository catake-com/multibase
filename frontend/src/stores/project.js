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
    openedProjectIDs: [0, 1],
    currentProjectID: 0,
  }),
  actions: {
    openGRPCProject(newProjectID, grpcProjectID) {
      if (this.openedProjectIDs.includes(grpcProjectID)) {
        this.openedProjectIDs = this.openedProjectIDs.filter((pID) => pID !== newProjectID);
      } else {
        this.openedProjectIDs = this.openedProjectIDs.map((projectID) => {
          return projectID === newProjectID ? grpcProjectID : projectID;
        });
      }

      this.currentProjectID = grpcProjectID;
      delete this.projects[newProjectID];

      this.saveState();
    },

    createNewGRPCProject(grpcProjectID) {
      this.projects[grpcProjectID] = {
        type: "grpc",
      };

      this.saveState();
    },

    createNewProject() {
      const projectID = Math.floor(Date.now() * Math.random());
      this.projects[projectID] = {
        type: "new",
      };
      this.openedProjectIDs.push(projectID);
      this.currentProjectID = projectID;

      this.saveState();
    },

    closeProjectTab(projectID) {
      if (this.openedProjectIDs.length <= 1) {
        return;
      }

      if (this.projects[projectID].type === "new") {
        delete this.projects[projectID];
      }

      this.openedProjectIDs = this.openedProjectIDs.filter((pID) => pID !== parseInt(projectID));

      this.currentProjectID = this.openedProjectIDs[0];

      this.saveState();
    },

    saveState() {
      const state = {
        projects: this.projects,
        openedProjectIDs: this.openedProjectIDs,
        currentProjectID: this.currentProjectID,
      };

      localStorage.setItem("projectState", JSON.stringify(state));
    },

    loadState() {
      const state = JSON.parse(localStorage.getItem("projectState")) || {};

      if ("projects" in state) {
        this.projects = state.projects;
      }

      if ("openedProjectIDs" in state) {
        this.openedProjectIDs = state.openedProjectIDs;
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
