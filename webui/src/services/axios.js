import axios from "axios";

const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});

// after a page refresh, put the saved token back in the Authorization header
let token = localStorage.getItem("token");
if (token) {
	instance.defaults.headers.common["Authorization"] = "Bearer " + token;
}

export default instance;
