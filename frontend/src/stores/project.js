import { defineStore } from "pinia";

import {
  State,
  CreateGRPCProject,
  CreateThriftProject,
  CreateNewProject,
  OpenProject,
  CloseProject,
  SaveCurrentProjectID,
} from "../wailsjs/go/project/Module";

export const useProjectStore = defineStore({
  id: "project",
  state: () => ({
    projects: {
      "404f5702-6179-4861-9533-b5ee16161c78": {
        type: "new",
      },
    },
    openedProjectIDs: ["404f5702-6179-4861-9533-b5ee16161c78"],
    currentProjectID: "404f5702-6179-4861-9533-b5ee16161c78",
  }),
  actions: {
    openProject(newProjectID, projectToOpenID) {
      OpenProject(newProjectID, projectToOpenID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    createNewGRPCProject(grpcProjectID) {
      CreateGRPCProject(grpcProjectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    createNewThriftProject(thriftProjectID) {
      CreateThriftProject(thriftProjectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    async createNewProject() {
      return CreateNewProject()
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    closeProjectTab(projectID) {
      CloseProject(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
        });
    },

    saveCurrentProjectID(projectID) {
      SaveCurrentProjectID(projectID)
        .then((state) => {
          this.$state = state;
        })
        .catch((reason) => {
          console.log(reason);
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
