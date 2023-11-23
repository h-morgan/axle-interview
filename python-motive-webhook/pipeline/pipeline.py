import json
import os

from clients.motive import MotiveClient
from loguru import logger


class MotivePipeline:
    def __init__(self, token: str, resource_name) -> None:
        self.motive_api = MotiveClient(token)
        self.resource_name = resource_name

        # these are the current resources we support for extraction from Motive
        self.extract_fn_map = {
            "drivers": self.motive_api.get_drivers,
            "vehicles": self.motive_api.get_vehicles,
            "trailers": self.motive_api.get_trailers,
        }

    def run(self) -> list:
        """
        Runs EL pipeline for Motive data:
        """

        # extract
        data = self.extract()

        # load
        metrics = self.load(data)

        return metrics

    def extract(self) -> list:

        data = self.extract_fn_map[self.resource_name]()
        logger.debug(f"extract data {data}")
        logger.info(f"{len(data)} {self.resource_name} extracted")

        return data

    def load(self, data) -> dict[str, int | str]:
        env = os.getenv("ENV", "dev")

        load_metadata = {
            "status": "success",
            "resource": self.resource_name,
            "num_items": len(data),
            "env": env,
            "location": "",
            "data": {self.resource_name: data}
            # TODO add dates, load time metadata, etc.
        }
        if env == "dev":
            logger.debug("Running in dev env, loading files locally")
            # create output dir if necessary
            output_dir = "data/"
            if not os.path.exists(output_dir):
                os.makedirs(output_dir)
                logger.debug(f"Output directory {output_dir} created")
            load_metadata["location"] = output_dir
            with open(f"{output_dir}/{self.resource_name}.json", "w") as fp:
                json.dump(data, fp)

        # TODO make work with s3
        logger.info(f"{len(data)} {self.resource_name} loaded")

        return load_metadata
