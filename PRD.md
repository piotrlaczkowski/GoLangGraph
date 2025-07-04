The goal of this project is to provide fully functional implementation of LangGraph Python framework for GoLang. Implement entire functionality of LangGraph and langgraph-cli (for easy deployment of agents). We need to be able to: 

- connect to various providers (local ollama, Gemini, OpenAI etc -> similarly to lightllm etc).
- easily define react_agent like it is availble in langgraph
- easily define graphs and states and debug them visually
- easily deploy the agents to miltiple roots in a simple way (the entire api, ops shoudl be somehow handled by the package on it's own)
- save the state of the graph (either in a database or in memory), communicate with databases (Postgres, local ones etc ...), handle sessions, threads, and everything we need 
- be able to define multiple collaborating agents and handle it's logic
- all other langgraph functionalities, with well defined USer/agent messages, messages history etc

TODO:

Make sure that we:

- handle all the react agent 
- handle  workflows efficiently
- handle state persistance, session threads and other parts of the langgraph
- have langgraph-cli functionality implemented correctly etc
- can handle mcp protocol easily
- can handle agents collaboration and advanced agets swarms configurations and communication protocols with tasks delegation etc 

to do all that you can check the github langgraph code implementation and reimplement everything in GoLang (so that we are really sure we have full functionality) [LangGraph GitHub](https://github.com/langchain-ai/langgraph)