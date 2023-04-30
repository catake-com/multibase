import { acceptHMRUpdate, defineStore } from "pinia";

import {
  CreateNewProject,
  SaveState,
  DeleteProject,
  ProjectState,
  Connect,
  Topics,
  Brokers,
  Consumers,
  StartTopicConsuming,
  StopTopicConsuming,
} from "../wailsjs/go/kafka/Module";

import { EventsOff, EventsOn } from "../wailsjs/runtime";

export const useKafkaStore = defineStore({
  id: "kafka",
  state: () => ({
    projectStates: {},
    consumersDataByProjectID: {},
    topicsDataByProjectID: {},
    brokersDataByProjectID: {},
    initiatedTopicConsumingByProjectID: {},
    consumedTopicsByProjectID: {},
    consumedTopicsMessagesByProjectID: {},
  }),
  getters: {
    projectState: (state) => {
      return (projectID) => state.projectStates[projectID] || {};
    },
    consumersData: (state) => {
      return (projectID) => state.consumersDataByProjectID[projectID] || {};
    },
    topicsData: (state) => {
      return (projectID) => state.topicsDataByProjectID[projectID] || {};
    },
    brokersData: (state) => {
      return (projectID) => state.brokersDataByProjectID[projectID] || {};
    },
    initiatedTopicConsuming: (state) => {
      return (projectID) => state.initiatedTopicConsumingByProjectID[projectID] || {};
    },
    consumedTopic: (state) => {
      return (projectID) => state.consumedTopicsByProjectID[projectID] || {};
    },
    consumedTopicMessages: (state) => {
      return (projectID) => state.consumedTopicsMessagesByProjectID[projectID] || {};
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

    async saveState(projectID, state) {
      try {
        this.projectStates[projectID] = await SaveState(projectID, state);
      } catch (error) {
        console.log(error);
      }
    },

    async loadTopics(projectID) {
      this.topicsDataByProjectID[projectID] = await Topics(projectID);
    },

    async loadBrokers(projectID) {
      this.brokersDataByProjectID[projectID] = await Brokers(projectID);
    },

    async loadConsumers(projectID) {
      this.consumersDataByProjectID[projectID] = await Consumers(projectID);
    },

    async connect(projectID) {
      this.projectStates[projectID] = await Connect(projectID);
    },

    initiateTopicConsuming(projectID, topic) {
      if (!this.consumedTopicsMessagesByProjectID[projectID]) {
        this.consumedTopicsMessagesByProjectID[projectID] = [];
      }

      this.initiatedTopicConsumingByProjectID[projectID] = { topicName: topic };
    },

    async startTopicConsuming(projectID, topic, startFromTime) {
      EventsOn(`kafka_message_${projectID}`, (data) => {
        this.consumedTopicsMessagesByProjectID[projectID].push(data);
      });

      this.consumedTopicsByProjectID[projectID] = await StartTopicConsuming(projectID, topic, startFromTime);
    },

    async stopTopicConsuming(projectID) {
      EventsOff(`kafka_message_${projectID}`);
      this.initiatedTopicConsumingByProjectID[projectID] = {};
      this.consumedTopicsByProjectID[projectID] = {};
      this.consumedTopicsMessagesByProjectID[projectID] = [];
      await StopTopicConsuming(projectID);
    },

    async restartTopicConsuming(projectID, startFromTime) {
      const currentTopic = this.initiatedTopicConsumingByProjectID[projectID].topicName;

      EventsOff(`kafka_message_${projectID}`);
      this.consumedTopicsByProjectID[projectID] = {};
      this.consumedTopicsMessagesByProjectID[projectID] = [];
      await StopTopicConsuming(projectID);

      await this.startTopicConsuming(projectID, currentTopic, startFromTime);
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
  import.meta.hot.accept(acceptHMRUpdate(useKafkaStore, import.meta.hot));
}
