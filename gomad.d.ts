// GOMAD Auto-Generated Definitions
// Generated at: 2025-11-27T12:02:14+03:00

export interface User {
    id: number;
    username: string;
    email: string;
    created_at: string;
    is_active: boolean;
}

export interface GomadAPI {
    call<T = any>(method: string, ...args: any[]): Promise<T>;

    // Binding: getVersion
    call(method: 'getVersion'): Promise<string>;

    // Binding: greet
    call(method: 'greet', arg0: string): Promise<string>;

    // Binding: add
    call(method: 'add', arg0: number, arg1: number): Promise<number>;

    // Binding: getUser
    call(method: 'getUser', arg0: number): Promise<User>;

    // Binding: divide
    call(method: 'divide', arg0: number, arg1: number): Promise<number>;

    // Binding: longTask
    call(method: 'longTask', arg0: number): Promise<string>;
}


declare global {
    interface Window {
        gomad: GomadAPI & {
            on(event: string, callback: (data: any) => void): () => void;
            off(event: string, callback?: (data: any) => void): void;
        };
    }
}