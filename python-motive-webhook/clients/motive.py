import requests
from loguru import logger
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry


class MotiveClient:
    """
    Sets up connection to Motive API and facilitates requests
    """

    def __init__(self, token: str) -> None:
        self.headers = {"Authorization": f"Bearer {token}"}
        self.base_url = "https://api.gomotive.com/v1/"

        self.page_size = 100

        # setup request session, retries, etc.
        session = requests.Session()
        retry = Retry(
            total=5, backoff_factor=1, status_forcelist=(500, 502, 503, 504, 429)
        )
        adapter = HTTPAdapter(max_retries=retry)
        session.mount("http://", adapter)
        session.mount("https://", adapter)
        self._session = session

    def query_api(self, endpoint: str, params: dict = {}) -> list:

        page_no = 1
        data = []
        # add pagination params to request with current page no
        params.update({"per_page": self.page_size, "page_no": page_no})
        while True:
            url = self.base_url + f"{endpoint}"

            # make request, raise exceptions if they come up
            logger.debug(f"Motive API request URL: {url}")
            logger.debug(f"Motive API request params: {params}")
            resp = self._session.get(url, headers=self.headers, params=params)
            resp.raise_for_status()

            logger.debug(f"Motive API response status code: {resp.status_code}")

            # get actual data piece from API response, key name in response conveniently same as endpoint name
            resp_json = resp.json()
            resp_data = resp_json.get(endpoint)
            logger.debug(
                f"Retrieved {len(resp_data)} {endpoint} from Motive API, on page {page_no}"
            )
            data.extend(resp_data)

            # handle pagination if necessary
            pagination = resp_json.get("pagination")
            if int(pagination["page_no"]) * int(pagination["per_page"]) >= int(
                pagination["total"]
            ):
                break
            else:
                page_no += 1

        return data

    def get_drivers(self) -> dict:

        endpoint = "users"
        params = {"role": "driver"}

        drivers = self.query_api(endpoint, params)
        return drivers

    def get_vehicles(self) -> dict:

        endpoint = "vehicles"
        drivers = self.query_api(endpoint)
        return drivers

    def get_trailers(self) -> dict:

        endpoint = "assets"
        drivers = self.query_api(endpoint)
        return drivers
