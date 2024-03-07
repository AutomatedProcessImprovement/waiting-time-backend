from flask import Blueprint, request, jsonify
import pandas as pd
from openai import OpenAI
import time
from datetime import datetime
import threading
import uuid
from io import StringIO
import requests
import app
import logging
import json
import os

logging.basicConfig(
    level=logging.INFO,  
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',  
    datefmt='%Y-%m-%d %H:%M:%S', 
    handlers=[
        logging.FileHandler('/app/logs/app.log'), 
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

chat_blueprint = Blueprint('chat', __name__)


openai_api_key = os.environ.get("OPENAI_API_KEY")
# Initialize the OpenAI client
client = OpenAI(
    api_key=openai_api_key
)

assistant = client.beta.assistants.retrieve("asst_wC8a4YpPfz7jWl24MkVwp0km")

instructions_table = "Use '{table_name}' as a placeholder for the actual table name in your query. In response time is given in seconds, so you need to convert it better (years/months/days/hours/minutes) format before providing an answer."
instructions = "Use function calling to get data about waiting times. Available columns: starttime, endtime, wtcontention, wtbatching, wtprioritization, wtunavailability, wtextraneous, sourceactivity, destinationactivity, sourceresource, destinationresource, wttotal. Waiting times are located in transitions between a pair of sequentially executed activities. Kronos identifies the duration of each execution of a transition. Then, it decomposes the duration into intervals by a waiting time cause. Kronos identifies intervals due to 5 causes: batching, prioritization, resource contention, resource unavailability, and extraneous factors. Example: when processing case ID12, there is a transition execution between activities A and B with a duration of 10 hours, decomposed as 1 hour - due to batching, 2 hours - due to prioritization, 3 hours - due to resource contention, 3 hours - due to resource unavailability, and 1 hour - due to extraneous factors. As a result of the analysis, Kronos compiles a report about the causes of waiting times for each transition execution.  The file appears to be a CSV (Comma-Separated Values) file containing the following columns: start_time: The start time of the transition. end_time: The end time of the transition. source_activity: The activity from which the case is transferred. source_resource: The resource associated with the source activity. destination_activity: The activity to which the case is transferred. destination_resource: The resource associated with the destination activity. case_id: The identifier of the business case. wt_total: The total waiting time for the transition execution. wt_contention: Waiting time due to resource contention. wt_batching: Waiting time due to batching. wt_prioritization: Waiting time due to prioritization. wt_unavailability: Waiting time due to resource unavailability. wt_extraneous: Waiting time due to extraneous factors."
base_url = "http://193.40.11.151"
added_text = "When presenting the results of the analysis, use a convenient, readable data format for waiting times (not units, but time in years, months, days, hours, minutes). Initial data is in seconds"

default_function_descriptions = {
    "discover_case_attributes": "Discovers and returns case attributes present in the event log. If returned data is empty, then there is no case attributes present in the event log.",
    "discover_batching_strategies": "Returns batching strategies of an event log. The function internally processes a pre-loaded event log and discovers batching strategies, providing insights into batch processing instances within a process. It characterizes each batch with details about the activity, involved resources, batch processing type (sequential, concurrent, or parallel), frequency, batch size distribution, duration distribution, and firing rules for batch initiation.",
    "discover_prioritization_strategies": "Discovers and returns prioritization strategies from an event log. This function analyzes the pre-loaded event log to identify case priority levels and classify process cases into corresponding levels. It provides insights into how cases are prioritized, ensuring that high-priority activities are executed before lower-priority ones when enabled simultaneously. The function categorizes cases into priority levels (e.g., low, medium, high) and identifies the rules that determine the priority level of each case.",
    "get_redesign_pattern_info": "Provides detailed information about specific redesign patterns. The information includes definitions, explanations, examples, positive impacts, negative impacts, and references. The function accepts one or several pattern names and returns comprehensive details about each.",
    "discover_resource_calendars": "Discovers and returns list of resources and their working calendars"
}

default_tools = [
    {
        "function": {
            "name": "discover_batching_strategies",
            "description": default_function_descriptions["discover_batching_strategies"],
            "parameters": {},
        },
        "type": "function",
    },
    {
        "function": {
            "name": "discover_resourse_calendars",
            "description": default_function_descriptions["discover_resource_calendars"],
            "parameters": {},
        },
        "type": "function",
    },
    {
        "function": {
            "name": "discover_prioritization_strategies",
            "description": default_function_descriptions["discover_prioritization_strategies"],
            "parameters": {},
        },
        "type": "function",
    },
    {
        "function": {
            "name": "discover_case_attributes",
            "description": default_function_descriptions["discover_case_attributes"],
            "parameters": {},
        },
        "type": "function",
    },
    {
        "function":{
            "name":"get_redesign_pattern_info",
            "description":"Provides detailed information about specific redesign patterns. The information includes definitions, explanations, examples, positive impacts, negative impacts, and references. The function accepts one or several pattern names and returns comprehensive details about each.",
            "parameters":{
                "type": "object",
                "properties": {
                    "patterns": {
                        "type": "array",
                        "description": "The list of redesign pattern names for which information is requested. Possible patterns: Contact reduction, First-contact problem resolution, Follow-up, Customer integration, Customer scheduling, Arrival time incentives, Task elimination, Fragment elimination, Task composition (combination), Task decomposition, Process decomposition, Process standardization, Process generalization, Process centralization, Case-based work, Case buffering, Periodic action, Resequencing, Parallelism, Order types (Case types), Triage, Exception, Extra resources, Resource scheduling, Assign cases, Customer teams, Fixed assignment , Flexible assignment, Case reassignement, Split responsibilities, Specialist, Empower, Department-based assignment, Experience-based task assignment, Expertise-based task assignment, Performance-based task assignment, Role-based task assignment, Teamwork-based assignment, Workload-based task assignment, Task delegation, Inventory buffering, Data elimination, Data composition, Data standardization, Capture data at source, Buffer information, Task automation, Automate for environmental impact, Fragment automation, Process automation, Integral technology , Establish standardized interfaces, Batch strategy optimization, Prioritization strategy optimization, Resource schedule optimization.",
                        "items": {
                            "type": "string"
                        }
                    }
                },
            "required": ["patterns"]
            }
        },
        "type":"function"
    },
    {
        "function": {
            "name": "query_database",
            "description": "Executes a read-only SQL query on a specific database that contains analysis on waiting times and returns the results. The query should be limited to select statements. Use '{table_name}' as a placeholder for the actual table name in your query. Available columns: starttime, endtime, wtcontention, wtbatching, wtprioritization, wtunavailability, wtextraneous, sourceactivity, destinationactivity, sourceresource, destinationresource, wttotal.",
            "parameters": {
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "The SQL query to execute. Include '{table_name}' as a placeholder for the table name."
                    }
                },
                "required": ["query"]
            },
        },
        "type": "function"
    }
]

def create_thread(jobid, message):
    thread = client.beta.threads.create()
    return thread

@chat_blueprint.route('/start', methods=['POST'])
def start_request():
    try:
        data = request.json
        jobid = data['jobid']
        message = data['message']
        instructions_req = data['instructions']
        assistant_instructions = data.get('assistantInstructions', '')
        selected_model = data['model']
        updated_descriptions = data.get('functionDescriptions', {})
        function_enabled_flags = data.get('functionStatus', {})

        logger.info("Enabled functions: " + str(function_enabled_flags))
        logger.info("Updated descriptions" + str(updated_descriptions))

        # Prepare the update parameters
        update_params = {
            "assistant_id": assistant.id,
            "model": selected_model,
            "tools": []
        }

        # Only include instructions if they are not empty
        # if assistant_instructions.strip():
        #     update_params["instructions"] = assistant_instructions

        # Update tools based on the request
        for tool in default_tools:
            function_name = tool["function"]["name"]
            # Ensure 'query_database' is always included
            if function_name == 'query_database' or function_name == "discover_resource_calendars" or function_enabled_flags.get(function_name, True):
                # Get the tool description, either the updated one or the default
                tool_description = updated_descriptions.get(function_name, tool["function"]["description"])
                tool["function"]["description"] = tool_description
                update_params["tools"].append(tool)

        logger.info(f"Update params: {update_params}")
        logger.info(f"Updated??? {str(client.beta.assistants.update(**update_params))}")

        new_thread = create_thread(jobid, message)
        logger.info(f"New thread created: {new_thread.id}")

        message_response = client.beta.threads.messages.create(
            thread_id=new_thread.id,
            role="user",
            content=message,
        )
        logger.info(f"Message sent in new thread {new_thread.id} with instructions: {instructions_req}")

        run_response = client.beta.threads.runs.create(
            thread_id=new_thread.id,
            assistant_id=assistant.id,
            instructions=instructions_req
        )
        logger.info(f"Run created in thread {new_thread.id}")

        return jsonify({'thread_id': new_thread.id, 'run_id': run_response.id})
    except Exception as e:
        logger.error(f"Error in start_request: {str(e)}")
        return jsonify({'error': str(e)}), 500


@chat_blueprint.route('/process', methods=['POST'])
def process_request():
    try:
        logger.info(f"Message sent in existing thread!")

        data = request.json
        threadid = data['threadId']
        jobid = data['jobid']
        message = data['message']
        instructions_req = data['instructions']
        
        logger.info(f"Message sent in existing thread with params {threadid}, {jobid}, {message}, {instructions_req}")

        message_response = client.beta.threads.messages.create(
            thread_id=threadid,
            role="user",
            content=message+added_text,
        )
        logger.info(f"Message sent in existing thread {threadid}")

        run_response = client.beta.threads.runs.create(
            thread_id=threadid,
            assistant_id=assistant.id,
            instructions=instructions_req
        )
        logger.info(f"Run created in existing thread {threadid}")

        return jsonify({'thread_id': threadid, 'run_id': run_response.id})
    except Exception as e:
        logger.error(f"Error in process_request: {str(e)}")
        return jsonify({'error': str(e)}), 500


@chat_blueprint.route('/status/<jobid>/<threadid>/<runid>', methods=['GET'])
def message_status(jobid, threadid, runid):
    try:
        run = client.beta.threads.runs.retrieve(
            thread_id=threadid,
            run_id=runid
        )
        run_status = run.status
        logger.info(f"Run status retrieved for thread {threadid}, run {runid}: {run_status}")

        if run_status == 'completed':
            messages = client.beta.threads.messages.list(threadid)
            last_five_messages = messages.data[:10]
            for msg in last_five_messages:
                logger.info(f"Message in thread {threadid}: {msg.content[0].text.value if msg.content else 'No content'}")
            response_message = messages.data[0].content[0].text.value
            logger.info(f"Response message for thread {threadid}, run {runid}: {response_message}")

            if response_message:
                return jsonify({'status': 'completed', 'message': response_message})

        if run_status == 'failed':
            return jsonify({'error': "Run has failed, problem on openai api side"}), 500

        tool_outputs = []
        if run_status == 'requires_action':
            logger.info(f"Processing required actions for thread {threadid}, run {runid}")
            for tool_call in run.required_action.submit_tool_outputs.tool_calls:
                tool_call_id = tool_call.id
                function_name = tool_call.function.name
                
                logger.info(f"Processing tool: {function_name} with tool_call_id: {tool_call_id}")
                
                # Initialize function_result
                function_result = None
                
                # Process based on function_name
                if function_name == 'query_database':
                    try:
                        function_arguments = json.loads(tool_call.function.arguments)
                        sanitized_jobid = app.DBHandler.sanitize_table_name(jobid)
                        table_name = f"result_{sanitized_jobid}"
                        query_template = function_arguments["query"]
                        query_template = query_template.replace("'{'table_name'}'", "{table_name}").replace('"{table_name}"', "{table_name}")
                        query_string = query_template.replace("{table_name}", table_name)
                        logger.info(f"Executing query: {query_string}")
                        query_result = app.execute_sql_query(jobid, query_string)
                        function_result = json.dumps(query_result)
                        logger.info(f"Query executed successfully, results obtained {function_result}")
                    except json.JSONDecodeError as json_err:
                        logger.error(f"JSON parsing error: {str(json_err)}")
                elif function_name == 'discover_batching_strategies':
                    function_result = str(app.perform_batching_discovery(jobid))
                    logger.info(f"Function {function_name} executed, results obtained {function_result}")
                elif function_name == 'discover_prioritization_strategies':
                    function_result = str(app.perform_prioritization_discovery(jobid))
                    logger.info(f"Function {function_name} executed, results obtained {function_result}.")
                elif function_name == 'discover_case_attributes':
                    function_result = str(app.case_attributes_discovery(jobid))
                    logger.info(f"Function {function_name} executed, results obtained.")
                elif function_name == 'get_redesign_pattern_info':
                    function_arguments = json.loads(tool_call.function.arguments)
                    function_result = str(get_redesign_pattern_info(function_arguments["patterns"]))
                    logger.info(f"Redesign pattern info retrieved successfully. {function_result}")
                elif function_name == 'discover_resourse_calendars':
                    function_result = str(app.perform_calendar_discovery(jobid))
                    logger.info(f"Function {function_name} executed, results obtained {function_result}.")
                
                if function_result is not None:
                    tool_outputs.append({
                        "tool_call_id": tool_call_id,
                        "output": function_result,
                    })
                    logger.info(f"Tool output for tool_call_id {tool_call_id} appended successfully.")

            if tool_outputs:
                run = client.beta.threads.runs.submit_tool_outputs(
                    thread_id=threadid,
                    run_id=runid,
                    tool_outputs=tool_outputs,
                )
                logger.info(f"Updated run status for thread {threadid}, run {runid}: {run.status}")
            if function_name == 'query_database':
                return jsonify({'status': run_status, 'message': f'{function_name}', 'function_result': f'{function_result}', 'function_arguments': f'{function_arguments["query"]}', 'function_name': f'{function_name}'})
            return jsonify({'status': run_status, 'message': f'{function_name}', 'function_result': f'{function_result}', 'function_name': f'{function_name}'})
        return jsonify({'status': run_status, 'message': ''})

    except Exception as e:
        logger.error(f"Error in message_status: {str(e)}")
        return jsonify({'status': "error", 'message': str(e)}), 500





def get_redesign_pattern_info(patterns):
    try:
        # Load the JSON data from the file
        with open('patterns.json', 'r') as json_file:
            redesign_patterns_info = json.load(json_file)
        
        # Fetch information for the requested patterns
        patterns_info = {pattern: redesign_patterns_info[pattern] for pattern in patterns if pattern in redesign_patterns_info}

        # Check if information was found for all requested patterns
        if len(patterns_info) != len(patterns):
            missing_patterns = set(patterns) - set(patterns_info.keys())
            available_patterns = list(redesign_patterns_info.keys())
            return ({'error': f'Information not found for patterns: {", ".join(missing_patterns)}',
                            'available_patterns': available_patterns}), 404

        return patterns_info, 200
    except Exception as e:
        return ({'error': str(e)}), 500















# def create_case_table(jobid):
#     # Fetch the event log CSV using the provided jobid
#     event_log_url = f"{base_url}/assets/results/{jobid}/event_log.csv"
#     event_log_response = requests.get(event_log_url)
    
#     if event_log_response.status_code != 200:
#         print(f"Failed to retrieve event log CSV for jobid {jobid}")
#         return None

#     event_log_csv = event_log_response.text
    
#     # Load the event log CSV into a pandas DataFrame
#     event_log_df = pd.read_csv(StringIO(event_log_csv))
    
#     # Fetch the column mapping for the job
#     jobs_url = f"{base_url}/jobs"
#     response = requests.get(jobs_url)
    
#     if response.status_code != 200:
#         print(f"Failed to retrieve job data for jobid {jobid}")
#         return None

#     job_data = response.json()
    
#     # Find the job with the given jobid
#     job = None
#     for j in job_data["jobs"]:
#         if j["id"] == jobid:
#             job = j
#             break

#     if job is None:
#         print(f"Job with jobid {jobid} not found")
#         return None
    
#     column_mapping = job["column_mapping"]
    
#     # Ensure the required columns are present in the event log DataFrame based on the column mapping
#     required_columns = [column_mapping["case"], column_mapping["start_timestamp"], column_mapping["end_timestamp"], column_mapping["activity"]]
#     for col_name in required_columns:
#         if col_name not in event_log_df.columns:
#             print(f"Column '{col_name}' not found in the event log.")
#             return None
    
#     # Group the event log entries by case_id and find the earliest start_time, latest end_time, and total processing time for each case
#     case_table = event_log_df.groupby(column_mapping["case"]).agg({
#         column_mapping["start_timestamp"]: "min",
#         column_mapping["end_timestamp"]: "max",
#         column_mapping["activity"]: ["count", "nunique"]
#     }).reset_index()
    
#     # Rename the columns in the case table
#     case_table.columns = ["case", "start_timestamp", "end_timestamp", "num_activities", "num_unique_activities"]
    
#     # Calculate the number of repeated executions of activities
#     case_table["num_repeated_executions"] = case_table["num_activities"] - case_table["num_unique_activities"]
    
#     # Calculate the total processing time for each case in seconds
#     case_table["total_processing_time"] = (pd.to_datetime(case_table["end_timestamp"]) - pd.to_datetime(case_table["start_timestamp"])).dt.total_seconds()
    
#     # Fetch the event log transitions report CSV using the provided jobid
#     event_log_transitions_report_url = f"{base_url}/assets/results/{jobid}/event_log_transitions_report.csv"
#     event_log_transitions_report_response = requests.get(event_log_transitions_report_url)
    
#     if event_log_transitions_report_response.status_code == 200:
#         event_log_transitions_report_csv = event_log_transitions_report_response.text
#         transitions_report_df = pd.read_csv(StringIO(event_log_transitions_report_csv))
        
#         # Calculate total_waiting_time for each case (sum of wt_total for each case) in seconds
#         total_waiting_time = transitions_report_df.groupby(column_mapping["case"])["wt_total"].sum().reset_index()
#         total_waiting_time.columns = ["case", "total_waiting_time"]
        
#         # Calculate total_cycle_time (total_processing_time + total_waiting_time) for each case
#         case_table = pd.merge(case_table, total_waiting_time, on="case", how="left")
#         case_table["total_cycle_time"] = case_table["total_processing_time"] + case_table["total_waiting_time"]
        
#         # Calculate CTE (cycle time efficiency) for each case
#         case_table["cte"] = case_table["total_processing_time"] / case_table["total_cycle_time"]
        
#         # Calculate number of transitions, number of unique transitions, and number of repeated executions of transitions
#         transitions_counts = transitions_report_df.groupby(column_mapping["case"])["source_activity"].count().reset_index()
#         transitions_counts.columns = ["case", "num_transitions"]
#         transitions_unique_counts = transitions_report_df.groupby(column_mapping["case"])["source_activity"].nunique().reset_index()
#         transitions_unique_counts.columns = ["case", "num_unique_transitions"]
#         case_table = pd.merge(case_table, transitions_counts, on="case", how="left")
#         case_table = pd.merge(case_table, transitions_unique_counts, on="case", how="left")
#         case_table["num_repeated_transitions"] = case_table["num_transitions"] - case_table["num_unique_transitions"]
    
#     # Check for additional columns (attributes) and transfer them to the case table if they are consistent within each case
#     additional_columns = [col for col in event_log_df.columns if col not in required_columns]
    
#     for col in additional_columns:
#         # Check if the column values are consistent within each case
#         consistent_values = event_log_df.groupby(column_mapping["case"])[col].nunique() == 1
#         if consistent_values.all():
#             case_table[col] = event_log_df.groupby(column_mapping["case"])[col].first()

#     # Export the case table as a CSV file
#     case_table_csv = case_table.to_csv(index=False)
    
#     # You can save the CSV to a file or return it as needed
#     # For example, to save it to a file:
#     # with open("case_table.csv", "w") as file:
#     #     file.write(case_table_csv)
    
#     return case_table_csv

# def create_thread(jobid, message):
#     batching_strategies = app.batching_strategies(jobid)
#     case_table = create_case_table(jobid)

#     event_log_transitions_report_url = f"{base_url}/assets/results/{jobid}/event_log_transitions_report.csv"
#     event_log_transitions_report_response = (requests.get(event_log_transitions_report_url)).text
#     path_transitions = "event_log_transitions_report.csv"
#     with open(path_transitions, "w") as file:
#         file.write(event_log_transitions_report_response)

#     transitions_file = client.files.create(
#         file=open("event_log_transitions_report.csv", "rb"),
#         purpose='assistants'
#     )

#     thread = client.beta.threads.create(
#         messages=[
#             {
#                 "role": "user",
#                 "content": message,
#                 "file_ids": [transitions_file.id]
#             }
#         ]
#     )
#     return thread








# assistant = client.beta.assistants.retrieve("asst_qq5TsyJEYMIHKwu02u4pGvPz")
# thread = client.beta.threads.retrieve("thread_6iSLlnNa84rArVkLd1HuDdyM")

# chats = {}
# chats_lock = threading.Lock()

# @chat_blueprint.route('/process/<jobid>/<message>')
# def process_request(jobid, message):
#     chat_id = str(uuid.uuid4())
#     try:
#         # Submit a message to the OpenAI API and create a run
#         message_response = client.beta.threads.messages.create(
#             thread_id=thread.id,
#             role="user",
#             content=message,
#         )

#         print("Message created")

#         run_response = client.beta.threads.runs.create(
#             thread_id=thread.id,
#             assistant_id=assistant.id,
#             instructions="Waiting times are located in transitions between a pair of sequentially executed activities. Kronos identifies the duration of each execution of a transition. Then, it decomposes the duration into intervals by a waiting time cause. Kronos identifies intervals due to 5 causes: batching, prioritization, resource contention, resource unavailability, and extraneous factors. Example: when processing case ID12, there is a transition execution between activities A and B with a duration of 10 hours, decomposed as 1 hour - due to batching, 2 hours - due to prioritization, 3 hours - due to resource contention, 3 hours - due to resource unavailability, and 1 hour - due to extraneous factors. As a result of the analysis, Kronos compiles a report about the causes of waiting times for each transition execution. In this report, each row is a transition execution, i.e., when a case is transferred between a pair of activities.Thus, start_time and end_time refer to the start and end time of the transition execution.source_activity is the activity from which the case is transferred.source_resource is the resource that executes the source_activity.destination_activity is the activity to which the case is transferred.destination_resource is the resource that executes the destination_activity.case_id is the identification label of the case.The transition execution has a duration that is waiting time; it is wt_total.The duration wt_total is decomposed into intervals based on the cause of waiting time:wt_contention is the interval of waiting time due to resource contention;wt_batching is the interval of waiting time due to batching;wt_prioritization is the interval of waiting time due to prioritization;wt_unavailability is the interval of waiting time due to  resource unavailability;wt_extraneous is the interval of waiting time due to extraneous (external) factors.Data units of waiting times are seconds. When presenting the results of the analysis, use a convenient readable data format, e.g., 2 years 2 months 2 days 2 minutes.Each transition (as defined by a unique combination of source_activity and destination_activity) might have multiple executions (rows in the dataset), and we should consider the combined waiting time across all these executions to determine the transition's duration.  Here is an explanation of the causes of waiting time.Waiting time due to batching occurs when an activity instance waits for another activity instance to be enabled in order to be processed together as a batch.Waiting time due to resource contention is observed when an activity instance waits to be processed by an assigned resource that is busy processing other activity instances, following a first-in-first-out (FIFO) order.Waiting time due to prioritization}is identified when the assigned resource is busy with an activity instance that was prioritized over the waiting one (not executed in the FIFO order).Waiting time due to resource unavailability occurs when the assigned resource is unavailable (off duty) due to their working schedules.We discover the working schedules of each resource to compare resource calendars with the waiting times observed in the log.Waiting time due to extraneous factors covers waiting times caused by external effects that cannot be identified from the event log -- e.g., the resource is working on another process, fatigue effects, or context switches."
#         )

#         print("Run created")

#         with chats_lock:
#             chats[chat_id] = {
#                 'message_id': message_response.id,
#                 'run_id': run_response.id,
#                 'status': 'processing'
#             }

#         return jsonify({'chat_id': chat_id})

#     except Exception as e:
#         return jsonify({'error': str(e)}), 500


# @chat_blueprint.route('/status/<chat_id>')
# def message_status(chat_id):
#     with chats_lock:
#         chat = chats.get(chat_id, None)

#     if chat is None:
#         return jsonify({'status': 'chat not found'}), 404

#     try:
#         run_status = client.beta.threads.runs.retrieve(
#             thread_id=thread.id,
#             run_id=chat['run_id']
#         ).status

#         print("Run status - ", run_status)

#         if run_status == 'completed':
#             messages = client.beta.threads.messages.list(thread.id)
#             response_message = messages.data[0].content[0].text.value

#             print("Response message - ", response_message)

#             if response_message:
#                 return jsonify({'status': 'completed', 'message': response_message})
        
#         return jsonify({'status': run_status, 'message': ''})

#     except Exception as e:
#         return jsonify({'error': str(e)}), 500



