import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import ChatView from '../views/ChatView.vue'
import EditGroupView from '../views/EditGroupView.vue'
import NewGroupView from '../views/NewGroupView.vue'
import ProfileView from '../views/ProfileView.vue'


const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/login', component: LoginView},
		{path: '/', component: HomeView},
		{path: '/profile', component: ProfileView},
		{path: '/conversations/:id', component: ChatView},
		{path: '/groups/new', component: NewGroupView},
		{path: '/groups/:id', component: EditGroupView},
	]
})

router.beforeEach((to) => {
	if (to.path != "/login" && !sessionStorage.getItem("token")) return "/login";
});

export default router
