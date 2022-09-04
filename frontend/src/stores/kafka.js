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
  Connect,
  Topics,
  Brokers,
  Consumers,
  StartTopicConsuming,
  StopTopicConsuming,
} from "@/wailsjs/go/kafka/Module";

import { EventsOff, EventsOn } from "@/wailsjs/runtime";

export const useKafkaStore = defineStore({
  id: "kafka",
  state: () => ({
    main: {
      projects: {
        "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {
          currentTab: "",
          authMethod: "",
          address: "",
          isConnected: false,
        },
      },
    },
    session: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {
        currentTopic: "",
        hoursAgo: 1,
      },
    },
    topics: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {
        isConnected: false,
        count: 0,
        list: [{ name: "", partitionCount: 0, messageCount: 0 }],
      },
    },
    brokers: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {
        isConnected: false,
        count: 0,
        list: [{ id: 0, rack: "", host: "", port: 0 }],
      },
    },
    consumers: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {
        isConnected: false,
        count: 0,
        list: [{ name: "", state: "" }],
      },
    },
    consumedTopic: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": {
        countTotal: 0,
        countCurrent: 0,
        partitions: [{ id: 0, offsetTotalStart: 0, offsetTotalEnd: 0, offsetCurrentStart: 0, offsetCurrentEnd: 0 }],
      },
    },
    consumedTopicMessages: {
      "dfaf4dc4-5fd1-42bb-b1ed-79b1d653279e": [{ timestamp: null, partitionID: 0, offset: 0, key: "", data: "" }],
    },
  }),
  actions: {
    async createNewProject(projectID) {
      try {
        this.$state.main = await CreateNewProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async deleteProject(projectID) {
      try {
        this.$state.main = await DeleteProject(projectID);
      } catch (error) {
        console.log(error);
      }
    },

    async saveCurrentTab(projectID, currentTab) {
      try {
        this.$state.main = await SaveCurrentTab(projectID, currentTab);
      } catch (error) {
        console.log(error);
      }
    },

    async saveAddress(projectID, address) {
      try {
        this.$state.main = await SaveAddress(projectID, address);
      } catch (error) {
        console.log(error);
      }
    },

    async saveAuthMethod(projectID, authMethod) {
      try {
        this.$state.main = await SaveAuthMethod(projectID, authMethod);
      } catch (error) {
        console.log(error);
      }
    },

    async saveAuthUsername(projectID, authUsername) {
      try {
        this.$state.main = await SaveAuthUsername(projectID, authUsername);
      } catch (error) {
        console.log(error);
      }
    },

    async saveAuthPassword(projectID, authPassword) {
      try {
        this.$state.main = await SaveAuthPassword(projectID, authPassword);
      } catch (error) {
        console.log(error);
      }
    },

    async loadTopics(projectID) {
      this.$state.topics[projectID] = await Topics(projectID);
    },

    async loadBrokers(projectID) {
      this.$state.brokers[projectID] = await Brokers(projectID);
    },

    async loadConsumers(projectID) {
      this.$state.consumers[projectID] = await Consumers(projectID);
    },

    async connect(projectID) {
      this.$state.main = await Connect(projectID);
    },

    async startTopicConsuming(projectID, topic, hours) {
      if (!this.$state.session[projectID]) {
        this.$state.session[projectID] = {};
      }

      if (!this.$state.consumedTopicMessages[projectID]) {
        this.$state.consumedTopicMessages[projectID] = [];
      }

      this.$state.session[projectID].currentTopic = topic;
      this.$state.session[projectID].hoursAgo = hours;

      EventsOn(`kafka_message_${projectID}`, (data) => {
        this.$state.consumedTopicMessages[projectID].push(data);
      });

      this.$state.consumedTopic[projectID] = await StartTopicConsuming(projectID, topic, hours);
    },

    async stopTopicConsuming(projectID) {
      EventsOff(`kafka_message_${projectID}`);
      this.$state.consumedTopicMessages[projectID] = [];
      this.$state.session[projectID] = {};

      await StopTopicConsuming(projectID);
    },

    async restartTopicConsuming(projectID) {
      EventsOff(`kafka_message_${projectID}`);
      this.$state.consumedTopicMessages[projectID] = [];
      await StopTopicConsuming(projectID);

      EventsOn(`kafka_message_${projectID}`, (data) => {
        this.$state.consumedTopicMessages[projectID].push(data);
      });

      this.$state.consumedTopic[projectID] = await ConsumeTopic(
        projectID,
        this.$state.session[projectID].currentTopic,
        this.$state.session[projectID].hoursAgo
      );
    },

    async loadState() {
      try {
        this.$state.main = await State();
      } catch (error) {
        console.log(error);
      }
    },
  },
});
