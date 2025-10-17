/* Do not change, this code is generated from Golang structs */


export enum ComicType {
    preview = 0,
    published = 1,
}
export interface ComicConfig {
    name: string;
    description: string;
}
export interface Comic {
    id: string;
    user_id: string;
    created_at: number;
    updated_at: number;
    config?: ComicConfig;
}
export interface LoginRequest {
    email: string;
    password: string;
}
export interface LoginResponse {
    token: string;
}
export interface UserStatusResponse {
    user_id: string;
}
export interface User {
    user_id: string;
    email: string;
    token: string;
}