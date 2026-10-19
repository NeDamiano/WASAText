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
				let response = await this.$axios.put("/groups/" + this.$route.params.id + "/name", { name: this.groupName });
				this.conv.name = response.data.name;
				this.groupName = "";
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async setGroupPhoto(event) {
			this.errormsg = null;
			let form = new FormData();
			form.append("image", event.target.files[0]);
			try {
				let response = await this.$axios.put("/groups/" + this.$route.params.id + "/photo", form);
				this.conv.photo = response.data.photo;
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async searchUsers() {
			this.errormsg = null;
			try {
				let response = await this.$axios.get("/users", { params: { username: this.search } });
				this.users = response.data.users.filter(x => !this.conv.members.some(m => m.id == x.id));
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		async addMember(userId) {
			this.errormsg = null;
			try {
				await this.$axios.post("/groups/" + this.$route.params.id + "/members", { userId: userId });
				this.users = this.users.filter(x => x.id != userId);
				this.refresh();
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
			<h1 class="h2 mb-4">
				<img v-if="conv.photo" :src="photoUrl(conv.photo)" width="80" height="80" class="rounded-circle me-2">
				<span v-else class="d-inline-flex align-items-center justify-content-center rounded-circle bg-secondary text-white me-2" style="width: 80px; height: 80px">{{ conv.name[0] }}</span>
				{{ conv.name }}
			</h1>
			<h2 class="h6">Name</h2>
			<div class="input-group mb-4">
				<input v-model="groupName" class="form-control" placeholder="New group name" @keyup.enter="setGroupName">
				<button class="btn btn-outline-primary" @click="setGroupName">Rename</button>
			</div>
			<h2 class="h6">Photo</h2>
			<label class="btn btn-outline-secondary mb-4">
				Change group photo
				<input type="file" accept="image/*" hidden @change="setGroupPhoto">
			</label>
			<h2 class="h6">Members</h2>
			<ul class="list-group mb-4">
				<li v-for="u in conv.members" :key="u.id" class="list-group-item">{{ u.name }}</li>
			</ul>
			<h2 class="h6">Add member</h2>
			<div class="input-group mb-2">
				<input v-model="search" class="form-control" placeholder="Search users" @keyup.enter="searchUsers">
				<button v-if="users.length" class="btn btn-outline-secondary" title="Clear" @click="search = ''; users = []">
					<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#x"/></svg>
				</button>
				<button class="btn btn-outline-primary" title="Search" @click="searchUsers">
					<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#search"/></svg>
				</button>
				<button class="btn btn-outline-secondary" @click="search = ''; searchUsers()">Show all</button>
			</div>
			<ul class="list-group mb-4">
				<li v-for="u in users" :key="u.id" class="list-group-item list-group-item-action" @click="addMember(u.id)">Add {{ u.name }}</li>
			</ul>
			<hr>
			<button class="btn btn-outline-primary me-2" @click="$router.push('/conversations/' + conv.id)">Back to chat</button>
			<button class="btn btn-danger" @click="leaveGroup">Leave group</button>
		</div>
	</div>
</template>

<style>
</style>
