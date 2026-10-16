<script>
export default {
	data() { return { conv: null, groupName: "", search: "", users: [], errormsg: null } },
	methods: {
		async refresh() {
			try {
				let response = await this.$axios.get("/conversations/" + this.$route.params.id);
				this.conv = response.data;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async setGroupName() {
			this.errormsg = null;
			try {
				await this.$axios.put("/groups/" + this.$route.params.id + "/name", { name: this.groupName });
				this.groupName = "";
				this.refresh();
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async setGroupPhoto(event) {
			this.errormsg = null;
			let form = new FormData();
			form.append("image", event.target.files[0]);
			try {
				await this.$axios.put("/groups/" + this.$route.params.id + "/photo", form);
				this.refresh();
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async searchUsers() {
			this.errormsg = null;
			try {
				let response = await this.$axios.get("/users", { params: { username: this.search } });
				this.users = response.data.users;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async addMember(userId) {
			this.errormsg = null;
			try {
				await this.$axios.post("/groups/" + this.$route.params.id + "/members", { userId: userId });
				this.users = [];
				this.search = "";
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async leaveGroup() {
			this.errormsg = null;
			try {
				await this.$axios.delete("/groups/" + this.$route.params.id + "/members/me");
				this.$router.push("/");
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
		<div v-if="conv">
			<h1 class="h2 mb-3">
				<img v-if="conv.photo" :src="photoUrl(conv.photo)" width="40" height="40" class="rounded-circle me-2">
				{{ conv.name }}
			</h1>
			<div class="input-group mb-2">
				<input v-model="groupName" class="form-control" placeholder="New group name" @keyup.enter="setGroupName">
				<button class="btn btn-outline-primary" @click="setGroupName">Rename</button>
			</div>
			<label class="btn btn-outline-secondary mb-3">
				Change group photo
				<input type="file" accept="image/*" hidden @change="setGroupPhoto">
			</label>
			<div class="input-group mb-2">
				<input v-model="search" class="form-control" placeholder="Search users to add" @keyup.enter="searchUsers">
				<button class="btn btn-outline-primary" @click="searchUsers">Search</button>
			</div>
			<ul class="list-group mb-3">
				<li v-for="u in users" :key="u.id" class="list-group-item list-group-item-action" @click="addMember(u.id)">Add {{ u.name }}</li>
			</ul>
			<button class="btn btn-outline-primary me-2" @click="$router.push('/conversations/' + conv.id)">Back to chat</button>
			<button class="btn btn-danger" @click="leaveGroup">Leave group</button>
		</div>
	</div>
</template>

<style>
</style>
