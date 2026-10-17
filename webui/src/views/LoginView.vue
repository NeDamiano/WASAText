<script>
export default {
	data() {
		return {
			name: "",
			errormsg: null,
		}
	},
	methods: {
		async login() {
			this.errormsg = null;
			try {
				let response = await this.$axios.post("/session", { name: this.name });
				let id = response.data.identifier;
				sessionStorage.setItem("token", id);
				this.$axios.defaults.headers.common["Authorization"] = "Bearer " + id;
				this.$router.push("/");
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
	},
}
</script>

<template>
	<div class="card shadow-sm mx-auto mt-5" style="max-width: 400px">
		<div class="card-body p-4">
			<h1 class="h3 text-center">WASAText</h1>
			<p class="text-muted text-center mb-4">Enter your username to log in</p>
			<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
			<input v-model="name" class="form-control mb-3" placeholder="Username" @keyup.enter="login">
			<button class="btn btn-primary w-100" @click="login">Login</button>
		</div>
	</div>
</template>

<style>
</style>
