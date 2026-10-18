<script>
export default {
	data() { return { name: "", search: "", users: [], members: [], errormsg: null, me: Number(sessionStorage.getItem("token")) } },
	methods: {
		async searchUsers() {
			this.errormsg = null;
			try {
				let response = await this.$axios.get("/users", { params: { username: this.search } });
				this.users = response.data.users.filter(x => x.id != this.me);
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
		addMember(u) {
			if (!this.members.find(x => x.id == u.id)) this.members.push(u);
			this.users = this.users.filter(x => x.id != u.id);
		},
		removeMember(u) {
			this.members = this.members.filter(x => x.id != u.id);
			this.users.push(u);
		},
		async createGroup() {
			this.errormsg = null;
			try {
				let response = await this.$axios.post("/groups", { name: this.name, members: this.members.map(x => x.id) });
				this.$router.push("/conversations/" + response.data.id);
			} catch (e) {
				this.errormsg = e.response ? e.response.data.message : e.toString();
			}
		},
	},
}
</script>

<template>
	<div class="mt-3" style="max-width: 400px">
		<h1 class="h2 mb-4">New group</h1>
		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
		<h2 class="h6">Name</h2>
		<input v-model="name" class="form-control mb-4" placeholder="Group name">
		<h2 class="h6">Add members</h2>
		<div class="input-group mb-2">
			<input v-model="search" class="form-control" placeholder="Search users" @keyup.enter="searchUsers">
			<button class="btn btn-outline-primary" title="Search" @click="searchUsers">
				<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#search"/></svg>
			</button>
			<button class="btn btn-outline-secondary" @click="search = ''; searchUsers()">Show all</button>
		</div>
		<ul class="list-group mb-2">
			<li v-for="u in users" :key="u.id" class="list-group-item list-group-item-action" @click="addMember(u)">Add {{ u.name }}</li>
		</ul>
		<div class="mb-4">
			<span v-for="u in members" :key="u.id" class="badge text-bg-light border fw-normal fs-6 me-1 d-inline-flex align-items-center">
				{{ u.name }}
				<button class="btn-close ms-1" style="font-size: 0.5rem" @click="removeMember(u)"></button>
			</span>
		</div>
		<hr>
		<button class="btn btn-primary" @click="createGroup">Create group</button>
	</div>
</template>

<style>
</style>
