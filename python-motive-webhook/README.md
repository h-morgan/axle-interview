# python-motive-webhook

Since Python is my most comfortable language, I created a first pass of the service in Python. This has the same endpoints and expects the same inputs as the Go service.

I use poetry for a package manager here. To run this service, first run:

```bash
poetry install
```

This installs all needed dependencies, as defined in the pyproject.toml file.

Next, run:

```bash
poetry run python app.py
```

This will spin up a development server of the Flask app, running on http://127.0.0.1:8000.
