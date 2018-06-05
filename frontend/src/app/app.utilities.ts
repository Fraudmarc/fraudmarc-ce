type ISODateString = string;

export interface IDateRange {
    startDate: ISODateString;
    endDate: ISODateString;
}



export function directCopy(str) {
    document['oncopy'] = function (event) {
        event.clipboardData.setData('Text', str);
        event.preventDefault();
    };
    document.execCommand('Copy');
    document['oncopy'] = undefined;
}

export function sanitizePropertyNames(data: any, transform: Function = value => value) {
    Object.keys(data).forEach((prop) => {
        data[prop.replace(/-/g, '_')] = transform(data[prop]);
    });
}

export function sanitizePropertyNamesValues(data: any) {
    sanitizePropertyNames(data, (value) => {
        return (value === null) ? '' : value;
    });
}

export function getlast30DayRange() {
    const range: IDateRange = {} as IDateRange;
    const today = new Date();
    range.startDate = toMidnight(new Date(today.valueOf() - 2592000000), true);
    range.endDate = toMidnight(today);
    return range;
}

export function toMidnight(date: Date, am = false): ISODateString {
    // Take a Date clamp it to either 12:00:00:0000 or 23:59:59:999 and return 
    if (am) {
        date.setHours(0);
        date.setMinutes(0);
        date.setSeconds(0);
        date.setMilliseconds(0);
    } else {
        date.setHours(23);
        date.setMinutes(59);
        date.setSeconds(59);
        date.setMilliseconds(999);
    }
    return date.toISOString();
}