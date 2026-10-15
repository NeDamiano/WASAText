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
				localStorage.setItem("token", id);
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
	<div class="mt-5" style="max-width: 400px">
		<h1 class="h2 mb-3">Login</h1>
		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
		<input v-model="name" class="form-control mb-2" placeholder="Username">
		<button class="btn btn-primary" @click="login">Login</button>
	</div>
</template>

<style>
</style>
