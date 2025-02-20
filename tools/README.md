# OpenAI API Tool helper

This package can help you integrate the Tavily client as an OpenAI tool.

## Exemple

Tested against [Qwen2.5](https://qwen.readthedocs.io/en/latest/index.html) thru [vllm](https://qwen.readthedocs.io/en/latest/deployment/vllm.html) with [function calling](https://qwen.readthedocs.io/en/latest/framework/function_call.html#vllm) activated.

For reference:

```sh
docker run --rm --runtime nvidia --gpus all \
    -p 8000:8000 \
    --ipc=host \
    vllm/vllm-openai:latest \
    --model Qwen/Qwen2.5-72B-Instruct-AWQ \
    --served-model-name Qwen2.5-72B \
    --gpu-memory-utilization 0.99 \
    --enable-auto-tool-choice --tool-call-parser hermes
```

### Result

```raw
Question: What is Tavily? What do they offer? Be specific and do not omit anything. Have they made the news recently?
---------8<----------
Received 2 tool call request(s)
tool call: tavily_web_search {"query": "Tavily company information", "depth": "advanced", "result_format": "ranked", "category": "general"} chatcmpl-tool-4fc34f730a9c409ca21ed18af030fac3
tavily answer: [{<result><title>Tavily Company Profile 2024: Valuation, Funding & Investors - PitchBook</title><url>https://pitchbook.com/profiles/company/622135-63</url><score>0.999605</score><content>Tavily General Information Description. Developer of search engine technology designed for AI agents and tailored for Retrieval-Augmented Generation (RAG) purposes. The company's API helps connect LLMs and AI applications to trusted and real-time knowledge, thus reducing hallucinations and overall bias, enhancing their ability to make informed decisions, enabling data analysis, AI assistant</content></result> text} {<result><title>Tavily - Products, Competitors, Financials, Employees, Headquarters ...</title><url>https://www.cbinsights.com/company/tavily</url><score>0.998644</score><content>Tavily specializes in search engine technology for artificial intelligence applications within the tech industry. Use the CB Insights Platform to explore Tavily's full profile. Tavily - Products, Competitors, Financials, Employees, Headquarters Locations</content></result> text} {<result><title>Precision in AI Research: Tavily's Company Researcher</title><url>https://blog.tavily.com/companyresearcher/</url><score>0.998004</score><content>Precision in AI Research: Tavily’s Company Researcher Precision in AI Research: Tavily’s Company Researcher The Company Researcher automates a multi-stage workflow for real-time company analysis, integrating search and extraction with agentic behavior to generate high-quality results. Document Curation and Enrichment with Tavily Extract: Once a trusted cluster is identified, Tavily Extract pulls detailed data from these verified links, adding substantial depth to the research. Grounding is core to Tavily’s Company Researcher workflow. Tavily’s Company Researcher shows how you can utilize the strengths of each and have both tools work with one another to generate the results you want. By making simple adjustments, you can expand its use beyond company research to tackle various data-intensive tasks, ensuring consistent and reliable results.</content></result> text} {<result><title>Tavily - LinkedIn</title><url>https://www.linkedin.com/company/tavily</url><score>0.997933</score><content>Tavily is a search engine, specifically designed for AI agents and tailored for RAG purposes. This includes: 🎮 Advanced NPCs in gaming 🤖 Interactive customer service assistants 🗣️ Real-time communication avatars By integrating Tavily's web search capabilities, ACE Agents can access real-time information, enhancing their versatility and capabilities. Explore NVIDIA ACE Agents with Tavily to build more human-like conversational AI: 🔗 https://lnkd.in/dDdMN8Uf Find the full Langchain workflow here: 🔗 https://lnkd.in/d9uWvAbN 🚚 🔗 Transforming Supply Chain Risk Management with AI Agents Discover how Adaptive Web Search and ReAct Prompting can elevate risk management: 🔎 Dynamic Search Refinement: The agents adapt in real time, diving deeper as they encounter relevant information for the most current insights.</content></result> text} {<result><title>tavily-python/examples/company_information.py at master - GitHub</title><url>https://github.com/tavily-ai/tavily-python/blob/master/examples/company_information.py</url><score>0.997766</score><content>tavily-python/examples/company_information.py at master · tavily-ai/tavily-python · GitHub GitHub Copilot Write better code with AI Code Search Find more, search less Advanced Security Enterprise-grade security features GitHub Copilot Enterprise-grade AI features Search code, repositories, users, issues, pull requests... Search Saved searches To see all available qualifiers, see our documentation. company_information.py company_information.py tavily_client \= TavilyClient(api_key\=os.environ["TAVILY_API_KEY"]) print(f"Calling Tavily company retrieval with query: {query}...\n[Note]: Using 'advanced' may take up to 10 seconds.") context \= tavily_client.get_company_info(query\=query, search_depth\="advanced", max_results\=7) You are a business analyst tasked with collecting and summarizing company information and insights from a given context. - financials: If company is public, information from recent reports and stock performance. For example: company_name: None. print(f"Calling OpenAI to extract information...") © 2024 GitHub, Inc. Footer navigation</content></result> text}]

tool call: tavily_web_search {"query": "Tavily recent news", "depth": "basic", "result_format": "summary", "category": "news"} chatcmpl-tool-a52858f4cd9e4289988ca38286d64d3a
tavily answer: [{Recent news regarding Tavily indicates that Spain is refusing to handle US cargo due to concerns that weapons may be sent to Israel. This statement was made by US Federal Maritime Commissioner Carl Bentzel. text}]

---------8<----------
Response:
It seems there was a mix-up in the news results. The recent news article I found does not pertain to Tavily but rather to a situation involving Spain and US cargo. Let me provide you with the information I gathered about Tavily:

### Tavily Overview
**Description:**
Tavily is a developer of search engine technology designed for AI agents and tailored for Retrieval-Augmented Generation (RAG) purposes. Their API helps connect Large Language Models (LLMs) and AI applications to trusted and real-time knowledge, which reduces hallucinations and overall bias, enhances the ability to make informed decisions, and enables data analysis.

### Key Features
1. **Advanced NPCs in Gaming:**
   - Tavily's technology can be used to create more sophisticated and realistic non-player characters (NPCs) in video games by providing them with real-time information and context.

2. **Interactive Customer Service Assistants:**
   - Their search capabilities can enhance customer service chatbots and virtual assistants, making them more responsive and knowledgeable.

3. **Real-Time Communication Avatars:**
   - Tavily can be integrated into communication systems to create avatars that can access and use real-time information, improving the quality of interactions.

4. **Dynamic Search Refinement:**
   - AI agents using Tavily can adapt their search queries in real time, diving deeper into relevant information to provide the most current and accurate insights.

### Products
- **Tavily Extract:**
  - This tool pulls detailed data from trusted sources, enriching the research process and ensuring high-quality results.
- **Company Researcher:**
  - Automates a multi-stage workflow for real-time company analysis, integrating search and extraction with agentic behavior to generate comprehensive and precise insights.

### Applications
- **Supply Chain Risk Management:**
  - Tavily's technology can be used to elevate risk management processes by providing dynamic and adaptive web search capabilities.

### Integration Examples
- **NVIDIA ACE Agents:**
  - Tavily has been integrated with NVIDIA's ACE (App for Conversational AI) to build more human-like conversational AI, demonstrating the versatility of their technology.

### Additional Resources
- **LinkedIn:**
  - [Tavily on LinkedIn](https://www.linkedin.com/company/tavily)
- **GitHub:**
  - [Tavily Python Examples](https://github.com/tavily-ai/tavily-python)

### Recent News
I did not find any recent news articles specifically about Tavily. The news result I initially received was unrelated to the company. If you have any specific areas you'd like to know more about, feel free to ask!
```

### Code

See the [example source code](https://github.com/hekmon/tavily/blob/main/tools/example/main.go) is the isolated go package.
