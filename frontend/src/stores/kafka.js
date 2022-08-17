import { defineStore } from "pinia";

import {
  CreateNewProject,
  State,
  SaveCurrentTab,
  SaveAuthMethod,
  SaveAuthUsername,
  SaveAuthPassword,
  SaveAddress,
  DeleteProject,
} from "../wailsjs/go/kafka/Module";

export const useKafkaStore = defineStore({
  id: "kafka",
  state: () => ({
    projects: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {
        currentTab: "",
        authMethod: "",
        address: "",
      },
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

    async saveCurrentTab(projectID, currentTab) {
      try {
        this.$state = await SaveCurrentTab(projectID, currentTab);
      } catch (error) {
        console.log(error);
      }
    },

    async saveAddress(projectID, address) {
      try {
        this.$state = await SaveAddress(projectID, address);
      } catch (error) {
        console.log(error);
      }
    },

    async saveAuthMethod(projectID, authMethod) {
      try {
        this.$state = await SaveAuthMethod(projectID, authMethod);
      } catch (error) {
        console.log(error);
      }
    },

    async saveAuthUsername(projectID, authUsername) {
      try {
        this.$state = await SaveAuthUsername(projectID, authUsername);
      } catch (error) {
        console.log(error);
      }
    },

    async saveAuthPassword(projectID, authPassword) {
      try {
        this.$state = await SaveAuthPassword(projectID, authPassword);
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
