<script>
export default {
	data() { return { name: "", user: null, errormsg: null, me: Number(sessionStorage.getItem("token")) } },
	methods: {
		async refresh() {
			try {
				let response = await this.$axios.get("/users");
				this.user = response.data.users.find(x => x.id == this.me);
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async setMyUserName() {
			this.errormsg = null;
			try {
				let response = await this.$axios.put("/me/username", { name: this.name });
				this.user = response.data;
				this.name = "";
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async setMyPhoto(event) {
			this.errormsg = null;
			let form = new FormData();
			form.append("image", event.target.files[0]);
			try {
				let response = await this.$axios.put("/me/photo", form);
				this.user = response.data;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		photoUrl(photo) { return __API_URL__ + photo; },
	},
	mounted() { this.refresh(); },
}
</script>

<template>
	<div class="mt-3" style="max-width: 400px">
		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
		<h1 v-if="user" class="h2 mb-4">
			<img v-if="user.photo" :src="photoUrl(user.photo)" width="80" height="80" class="rounded-circle me-2">
			{{ user.name }}
		</h1>
		<h2 class="h6">Name</h2>
		<div class="input-group mb-4">
			<input v-model="name" class="form-control" placeholder="New username" @keyup.enter="setMyUserName">
			<button class="btn btn-outline-primary" @click="setMyUserName">Change name</button>
		</div>
		<h2 class="h6">Photo</h2>
		<label class="btn btn-outline-secondary">
			Change profile photo
			<input type="file" accept="image/*" hidden @change="setMyPhoto">
		</label>
	</div>
</template>

<style>
</style>
