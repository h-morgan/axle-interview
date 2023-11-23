import requests
from flask import Flask, jsonify, request

from pipeline.pipeline import MotivePipeline

app = Flask(__name__)

# data store for events and URLs of subscribers who subscribe to each event
# ideally this would be in external persistant data storage elsewhere
EVENTS = {
    "vehicles": ["https://eoww187fd6vl0sa.m.pipedream.net"],
    "drivers": ["https://eoww187fd6vl0sa.m.pipedream.net"],
    "trailers": ["https://eoww187fd6vl0sa.m.pipedream.net"],
}


@app.route("/")
def home():
    return jsonify({"status": "success", "msg": "we're up"}), 200


@app.route("/subscribers", methods=["GET"])
def subscribers():
    return jsonify({"status": "success", "data": EVENTS}), 200


@app.route("/subscribe", methods=["POST"])
def subscribe():
    params = request.get_json(force=True)

    # arg validation --
    # this endpoint expects 2 params - event and url. check if any are missing
    expected_args = {"event", "callback_url"}
    missing_args = []
    for arg in expected_args:
        if arg not in params:
            missing_args.append(arg)

    if missing_args:
        return (
            jsonify(
                {
                    "status": "error",
                    "msg": f"Missing expected args in POST request: {missing_args}",
                }
            ),
            400,
        )

    event = params["event"]
    url = params["callback_url"]
    existing_events = list(EVENTS.keys())
    if params["event"] not in existing_events:
        return (
            jsonify(
                {
                    "status": "error",
                    "msg": f"Invalid event - event must be one of the following: {existing_events}",
                }
            ),
            400,
        )

    # if we passed all validations, subscribe callback URL to requested event
    else:
        EVENTS[event].append(url)
        return (
            jsonify(
                {
                    "status": "success",
                    "msg": f"Successfully subscribed url {url} to event {event}",
                }
            ),
            200,
        )


@app.route("/motive-pipeline", methods=["POST"])
def motive_pipeline():
    params = request.get_json(force=True)

    # need API token from motive
    token = params.get("token")

    # validate that we got the API token
    if token is None:
        return (
            jsonify(
                {
                    "status": "error",
                    "msg": "Missing API token, required to retrieve Motive data",
                }
            ),
            400,
        )

    # if we got token, attempt to connect to Motive and run EL pipeline for each resource

    for resource, subscribers in EVENTS.items():
        pipeline = MotivePipeline(token, resource)
        metrics = pipeline.run()
        for subscriber in subscribers:
            # notify each subscriber of pipeline completion for resources they're subscribed to
            requests.post(subscriber, json=metrics)

    return (
        jsonify(
            {
                "status": "success",
                "msg": "Completed data load from Motive API for new customer",
            }
        ),
        200,
    )


def run():
    app.run(port=8000, debug=True)


if __name__ == "__main__":
    run()
