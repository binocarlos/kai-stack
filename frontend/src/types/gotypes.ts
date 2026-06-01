/* Do not change, this code is generated from Golang structs */


export enum ExampleType {
    preview = 0,
    published = 1,
}
export interface ExampleConfig {
    name: string;
    description: string;
}
export interface ExampleRecord {
    id: string;
    user_id: string;
    created_at: number;
    updated_at: number;
    config?: ExampleConfig;
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
    email: string;
    roles: string[];
}
export interface User {
    user_id: string;
    email: string;
    token: string;
}