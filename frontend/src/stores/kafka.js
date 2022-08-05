import { defineStore } from "pinia";

import { CreateNewProject, State, DeleteProject } from "../wailsjs/go/kafka/Module";

export const useKafkaStore = defineStore({
  id: "kafka",
  state: () => ({
    projects: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {},
    },
  }),
  actions: {
    async createNewProject(projectID) {
      try {
        this.$state = await CreateNewProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async deleteProject(projectID) {
      try {
        this.$state = await DeleteProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async loadState() {
      try {
        this.$state = await State();
      } catch (error) {
        console.log(error);
      }
    },
  },
});
