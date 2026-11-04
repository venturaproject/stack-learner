import { AxiosInstance } from 'axios';
import pathFor from '@/lib/app-routes';

declare global {
    interface Window {
        axios: AxiosInstance;
    }

    var route: typeof pathFor;
}
